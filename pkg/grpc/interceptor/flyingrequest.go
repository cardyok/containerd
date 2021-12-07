package interceptor

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sync"

	"google.golang.org/grpc"

	"github.com/containerd/containerd/errdefs"
)

var SoftBypassNewReq bool

const (
	// as "/runtime.v1alpha2.RuntimeService/StopPodSandbox"
	CriDistinctString = `^/runtime\.(.+)\.RuntimeService`
	// as "/containerd.services.content.v1.Content/Info"
	ContainerdDistinctString = `^/containerd\.service\.(.+)`
)

var criReg, containerdReg *regexp.Regexp

func init() {
	criReg = regexp.MustCompile(CriDistinctString)
	containerdReg = regexp.MustCompile(ContainerdDistinctString)
}

// RequestCountDecider is a user-provided function for deciding whether count a request is flying.
type RequestCountDecider func(ctx context.Context, fullMethodName string, servingObject interface{}) bool

// FlyingRequestCountInterceptor returns a new unary server interceptors that counts the flying requests.
func FlyingRequestCountInterceptor(decider RequestCountDecider, wq *sync.WaitGroup) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if SoftBypassNewReq && filterBypassRequest(info.FullMethod) {
			return nil, fmt.Errorf("service entering lame duck status,all request will bypass: %w", errdefs.ToGRPC(errdefs.ErrUnavailable))
		}
		if decider(ctx, info.FullMethod, info.Server) {
			wq.Add(1)
			defer wq.Done()
		}
		resp, err := handler(ctx, req)
		return resp, err
	}
}

// FlyingReqCountDecider decides whether this grpc request should be intercepted
func FlyingReqCountDecider(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
	return decideCriRequest(fullMethodName) || decideContainerdRequest(fullMethodName)
}

// filterBypassRequest filters which requests should be ignored when server enters lame duck status
func filterBypassRequest(fullMethodName string) bool {
	return containerdReg.MatchString(fullMethodName) || criReg.MatchString(fullMethodName)
}

// decideCriRequest decides if the request belongs to cri
func decideCriRequest(fullMethodName string) bool {
	methodName := path.Base(fullMethodName)
	switch methodName {
	case
		"RunPodSandbox",
		"StartPodSandbox",
		"StopPodSandbox",
		"RemovePodSandbox",
		"CreateContainer",
		"StartContainer",
		"StopContainer",
		"RemoveContainer",
		"PauseContainer",
		"UnpauseContainer":
		return true
	default:
		return false
	}
}

// decideCriRequest decides if the request belongs to containerd
func decideContainerdRequest(fullMethodName string) bool {
	methodName := path.Base(fullMethodName)
	switch methodName {
	case
		"Info",
		"Status",
		"List",
		"Stats",
		"Get":
		return false
	default:
		return true
	}
}
