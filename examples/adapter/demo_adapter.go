package adapter
//
// import (
//     "encoding/json"
//     "github.com/go-opener/ctxflow/layer"
//     "xxxxx.com/pkg/golib/v2/base"
//     "github.com/mitchellh/mapstructure"
// )
//
// type ApiClientAdapter struct {
//     layer.Flow
//     *base.ApiClient
// }
//
// func (entity *ApiClientAdapter) PreUse(args ...interface{}) {
//     if len(args) > 0 {
//         entity.ApiClient = args[0].(*base.ApiClient)
//     }
//     entity.Flow.PreUse(args...)
// }
//
// func (entity *ApiClientAdapter) HttpPost(path string, opts base.HttpRequestOptions) (*base.ApiResult, error) {
//     return entity.ApiClient.HttpPost(entity.GetContext(),path,opts)
// }
//
// func (entity *ApiClientAdapter) HttpGet(path string, opts base.HttpRequestOptions) (*base.ApiResult, error) {
//     return entity.ApiClient.HttpGet(entity.GetContext(),path,opts)
// }
//
// func (entity *ApiClientAdapter) DecodeResponse(res *base.ApiResult, output interface{}) (errno int, err error) {
//     var r base.DefaultRender
//     if err = json.Unmarshal(res.Response, &r); err != nil {
//         entity.LogWarnf("http response decode err, err: %s", res.Response)
//         return errno, err
//     }
//
//     errno = r.ErrNo
//     if r.ErrNo != 0 {
//         entity.LogErrorf( "http response code: %d", r.ErrNo)
//         return errno, err
//     }
//
//     if _, ok := r.Data.(map[string]interface{}); !ok {
//         return errno, nil
//     }
//
//     if err := mapstructure.Decode(r.Data, &output); err != nil {
//         entity.LogWarnf( "api call data decode error: %s", err.Error())
//         return errno, err
//     }
//
//     return errno, nil
// }
//
// func (entity *ApiClientAdapter) DecodeResponseSimple(res *base.ApiResult) (result interface{},errno int, err error) {
// 	var r base.DefaultRender
// 	if err = json.Unmarshal(res.Response, &r); err != nil {
// 		entity.LogErrorf("http response decode err, err: %s", res.Response)
// 		return nil, errno, err
// 	}
//
// 	errno = r.ErrNo
// 	if r.ErrNo != 0 {
// 		entity.LogErrorf( "http response code: %d", r.ErrNo)
// 		return nil,errno, err
// 	}
//
// 	return r.Data,errno, nil
// }
