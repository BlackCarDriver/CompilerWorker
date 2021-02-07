package rpcClient

import (
	"codeRunner"
	"fmt"

	"../config"

	"github.com/apache/thrift/lib/go/thrift"
)

// 监听thrift调用请求
func ServerThrift() {
	transport, err := thrift.NewTServerSocket(config.ServerConfig.ServerAddr)
	if err != nil {
		panic(err)
	}

	handler := &myCodeRunner{}
	processor := codeRunner.NewCodeRunnerProcessor(handler)

	transportFactory := thrift.NewTBufferedTransportFactory(8192)
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	server := thrift.NewTSimpleServer4(
		processor,
		transport,
		transportFactory,
		protocolFactory,
	)
	fmt.Println("thrift server running...")
	if err := server.Serve(); err != nil {
		panic(err)
	}
}
