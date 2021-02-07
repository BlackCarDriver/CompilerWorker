namespace go baseService

// 注册服务到simpleServer需要先继承这个服务
service baseService {
    string ping(1: string str)
}

// 共用响应体
struct CommomResp {
    1: i32 status
    2: string msg
    3: binary payload
}
