package goto_rpc

import proto "encoding/protobuf/proto"

type RpcServiceFunc func(IContext, proto.Message)
type RpcMessageFactoryFunc func()proto.Message

type MethodInfo struct {
	method			string
	service_func	RpcServiceFunc
	req_factory		RpcMessageFactoryFunc
	rsp_factory		RpcMessageFactoryFunc
}

type MethodMap map[string]*MethodInfo
