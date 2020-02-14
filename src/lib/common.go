package lib

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func OpenMysqlDB(host string, port int, user, pass, dbName string) (db *sql.DB, err error) {
	db, err = sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", user, pass, host, port, dbName))
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	db.SetConnMaxLifetime(60 * time.Second)
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(5)
	return
}

func OpenRedisDB(host string, port int, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: "", // no password set
		DB:       db, // use default DB
	})
	return client
}

func ExecCmd(ctx context.Context, name string, args []string, callback func(string)) (err error) {
	cmd := exec.CommandContext(ctx, name, args...)
	r, w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = w
	err = cmd.Start()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	var g sync.WaitGroup
	g.Add(1)
	ctxChild, _ := context.WithCancel(ctx)
	go func() {
		defer g.Done()
		bufferedSize := 4096
		readBuf := make([]byte, bufferedSize)
		for {
			select {
			case <-ctxChild.Done():
				break
			default:
			}
			size, err := r.Read(readBuf)
			if io.ErrClosedPipe == err || err != nil {
				break
			}
			if size == 0 {
				continue
			}
			line := strconv.Quote(string(readBuf[:size]))
			callback(line[1 : len(line)-1])
		}
	}()
	err = errors.WithStack(cmd.Wait())
	r.Close()
	w.Close()
	g.Wait()
	return
}
func RunContainer(ctx context.Context, sender chan<- string, containerName, imageName string, params []string, logEntry *logrus.Entry) {
	defer close(sender)

	var dockerParams []string
	dockerParams = append(dockerParams, "run")
	dockerParams = append(dockerParams, "--name="+containerName)
	// 使用bridge网络
	dockerParams = append(dockerParams, "--network=bridge")
	dockerParams = append(dockerParams, params...)
	dockerParams = append(dockerParams, imageName)
	logEntry.Info(dockerParams)
	ctxChild, _ := context.WithCancel(ctx)
	err := ExecCmd(ctxChild, "docker", dockerParams, func(line string) { sender <- line })
	if err != nil {
		logEntry.Error(err.Error())
		sender <- fmt.Sprintf(`{"status":"fail", "msg":%q}`, err.Error())
	} else {
		sender <- fmt.Sprintf(`{"status":"success", "msg":"merge success"}`)
	}
	select {
	case <-ctx.Done():
		err = KillContainer(containerName, logEntry)
		if err != nil {
			logEntry.Error(err.Error())
		}
	default:
		cmd := exec.Command("docker", "rm", "-v", containerName)
		err = cmd.Run()
		if err != nil {
			logEntry.Error(err.Error())
			sender <- err.Error()
		}
	}
	return
}
func KillContainer(name string, logEntry *logrus.Entry) (err error) {
	logEntry.Infof("kill container:%s", name)
	cmd := exec.Command("docker", "rm", "-f", "-v", name)
	err = cmd.Run()
	return
}

func TemplateRender(templ string, data interface{}) (res string, err error) {
	t, err := template.New("").Parse(templ)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	res = buf.String()
	return
}
