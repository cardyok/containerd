/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package hotfix

import (
	"context"

	ptypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"

	api "github.com/containerd/containerd/api/services/hotfix/v1"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/plugin"
)

var _ api.HotfixServer = &service{}

func init() {
	plugin.Register(&plugin.Registration{
		Type:   plugin.GRPCPlugin,
		ID:     "hotfix",
		InitFn: initFunc,
	})
}

func initFunc(ic *plugin.InitContext) (interface{}, error) {
	return &service{}, nil
}

type service struct {
}

func (s *service) Register(server *grpc.Server) error {
	api.RegisterHotfixServer(server, s)
	return nil
}

func (s *service) ChangeLogLevel(ctx context.Context, request *api.ChangeLogLevelRequest) (*ptypes.Empty, error) {
	log.G(ctx).Infof("ChangeLogLevel: %s", request.LogLevel)
	return &ptypes.Empty{}, log.SetLevel(request.LogLevel)
}
