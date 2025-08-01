package notifier

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-logr/logr"
	"golang.org/x/net/http2"
	"k8s.io/client-go/util/workqueue"
)

type Request struct {
	Sid      string `json:"sid,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type task struct {
	req    Request
	method string
}

// Waiter 用于等待一个Workspace的Pod处于Ready状态
type Waiter interface {
	WaitFor(context.Context, string) error
}

// Notifier 用于通知一个Workspace可用（即它的Pod处于Ready状态）
// 注册或注销Workspace的IP地址到网关中，使得网关可以发现可用的Workspace
type Notifier interface {
	Login(sid, endpoint string)

	Logout(sid string)

	Notify(sid string)
}

type WorkspaceNotifier struct {
	logger logr.Logger
	// 通过HTTP请求来从网关中注册或注销Workspace
	clients []*http.Client
	Url     string
	Token   string

	ctx   context.Context
	queue workqueue.Interface

	mux sync.Mutex
	// 保存sid到chan的映射，用于通知指定的Workspace已经可用或等待Workspace可用
	wsc map[string]chan struct{}
}

func NewWorkspaceNotifier(ctx context.Context, logger logr.Logger, svcName, path, token string, workers int) (*WorkspaceNotifier, error) {
	// https://servicename/internal/endpoint
	w := &WorkspaceNotifier{
		logger: logger,
		Url:    fmt.Sprintf("https://%s%s", svcName, path),
		Token:  token,
		ctx:    ctx,
		queue:  workqueue.NewRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter()),
		wsc:    make(map[string]chan struct{}),
	}

	if workers <= 0 {
		workers = 4
	}

	go func() {
		<-ctx.Done()
		w.queue.ShutDown()
	}()

	// 开启http2
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 不校验服务端证书
	}
	err := http2.ConfigureTransport(transport)
	if err != nil {
		logger.Error(err, "configure transport")
		return nil, err
	}

	for i := 0; i < workers; i++ {
		w.clients = append(w.clients, &http.Client{
			Transport: transport,
		})
		go w.worker(i)
	}

	return w, nil
}

// Login 通过HTTP请求将Pod的IP地址和端口注册到网关中
// 使得网关可以访问到Pod
func (w *WorkspaceNotifier) Login(sid, endpoint string) {
	w.queue.Add(task{
		req:    Request{Sid: sid, Endpoint: endpoint},
		method: http.MethodPost,
	})
}

// Logout 从网关中注销Pod，防止网关访问到不存在或其它用户的Pod
func (w *WorkspaceNotifier) Logout(sid string) {
	w.queue.Add(task{
		req:    Request{Sid: sid},
		method: http.MethodDelete,
	})
}

func (w *WorkspaceNotifier) worker(i int) {
	client := w.clients[i]
	for {
		item, shutdown := w.queue.Get()
		if shutdown {
			return
		}

		tsk := item.(task)
		err := w.doRequest(client, tsk.req, tsk.method)
		if err != nil {
			w.logger.Error(err, "do request", "method", tsk.method, "sid", tsk.req.Sid)
		} else {
			w.queue.Done(item)
		}
	}
}

func (w *WorkspaceNotifier) doRequest(client *http.Client, req Request, method string) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(data)
	request, err := http.NewRequest(method, w.Url, reader)
	if err != nil {
		return err
	}

	request.Header.Set("token", w.Token)

	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}

// WaitFor 等待Pod可用
func (w *WorkspaceNotifier) WaitFor(ctx context.Context, sid string) error {
	w.logger.V(5).Info("WaitFor ", "sid", sid)
	var (
		ch chan struct{}
		ok bool
	)

	w.mux.Lock()
	ch, ok = w.wsc[sid]
	if !ok {
		w.logger.V(5).Info("creating chan", "sid", sid)
		ch = make(chan struct{})
		w.wsc[sid] = ch
	} else {
		w.logger.V(5).Info("find chan", "sid", sid)
	}
	w.mux.Unlock()

	defer func() {
		w.mux.Lock()
		w.logger.V(5).Info("deleting chan", "sid", sid)
		delete(w.wsc, sid)
		w.mux.Unlock()
	}()

	w.logger.V(5).Info("waiting for", "sid", sid)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
	}

	return nil
}

// Notify 通知Pod已经可用
func (w *WorkspaceNotifier) Notify(sid string) {
	w.logger.V(5).Info("Notify ", "sid", sid)

	w.mux.Lock()
	ch, ok := w.wsc[sid]
	w.mux.Unlock()
	if !ok {
		w.logger.V(5).Info("cant find chan", "sid", sid)
		return
	}
	w.logger.V(5).Info("closed chan", "sid", sid)
	close(ch)
}
