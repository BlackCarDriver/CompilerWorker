namespace go codeRunner

include "base.thrift"

/*
Thrift version:
    0.13.0
generate thrift:
    thrift -r --gen go -out D:\WorkPlace\GoWorkPlace\thriftGen\src codeRunner.thrift
*/


service codeRunner extends base.baseService {
    base.CommomResp buildGo(1: string code, 2: string input),
    base.CommomResp buildCpp(1: string code, 2: string input),
    base.CommomResp run(1: string ctype, 2: string hash, 3: string input),
}