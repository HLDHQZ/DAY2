package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"os/signal"
)

/*基于 errgroup 实现一个 http server 的启动和关闭 ，
以及 linux signal 信号的注册和处理，
要保证能够一个退出，全部注销退出*/
/*实现 HTTP server 的启动和关闭
监听 linux signal信号，支持 kill -9 或 Ctrl+C 的中断操作操作
errgroup 实现多个 goroutine 的级联退出*/


func StatHttpServer(srv *http.Server) error{
	http.HandleFunc("/hello",HelloServer)
	fmt.Println("http server start")
	err :=srv.ListenAndServe()
	return err
}

func HelloServer(w http.ResponseWriter,req *http.Request){
	io.WriteString(w,"hello,world!\n")
}

func main(){
	ctx := context.Background()

	ctx,cancel := context.WithCancel(ctx)

	group, errCtx := errgroup.WithContext(ctx)

	srv := &http.Server{Addr:":9090"}

	group.Go(func()error{
		return StatHttpServer(srv)
	})

	group.Go(func()error{
		<- errCtx.Done()
		fmt.Println("http server stop")
		return srv.Shutdown(errCtx)
	})

	chanel := make(chan os.Signal,1)
	signal.Notify(chanel)

	group.Go(func()error{
		for{
			select{
			case <- errCtx.Done():
				return errCtx.Err()
			case <- chanel:
				cancel()
			}
		}
		return nil
	})

	if err := group.Wait();err != nil{
		fmt.Println("group error:",err)
	}
	fmt.Println("all group done!")
}