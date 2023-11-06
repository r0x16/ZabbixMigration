package zabbix

import (
	"fmt"
	"net/http"
	"time"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"github.com/imroc/req/v3"
)

/**
 * ZabbixConnector is a struct that implements the Zabbix API
 * It uses the req library to make requests to the Zabbix API
 * It also uses the model package to handle the request and response
 * It also uses the Error model to handle errors
 */
type ZabbixConnector struct {
	Url    string
	client *req.Client
	Token  string
}

/**
 * API is a function that returns a ZabbixConnector
 * It receives a string with the URL of the Zabbix API
 * It returns a pointer to a ZabbixConnector
 */
func API(url string) *ZabbixConnector {

	connector := &ZabbixConnector{
		Url: url,
	}
	connector.initializeClient()

	return connector
}

/**
 * Connect is a function that connects to the Zabbix API
 * Sets the token in the ZabbixConnector for future requests
 */
func (z *ZabbixConnector) baseConnect(params model.ZabbixParams) *model.Error {
	body := z.UnauthorizedBody("user.login", params)

	response, err := z.Request(body)

	if err != nil {
		return err
	}

	z.Token = response.Result.(string)
	return nil
}

/**
 * Request is a function that makes a request to the Zabbix API
 * It receives a pointer to a ZabbixRequest
 * It returns a pointer to a ZabbixResponse and a pointer to an Error
 */
func (z *ZabbixConnector) Request(body *model.ZabbixRequest) (*model.ZabbixResponse, *model.Error) {
	var response model.ZabbixResponse

	resp := z.client.Post().SetBody(body).Do()

	if !resp.IsSuccessState() {
		return nil, &model.Error{
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("Connection Error %s", resp.Status),
			Data:    resp,
		}
	}

	err := resp.Into(&response)

	if err != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Data:    resp,
		}
	}

	if response.Error.Code != 0 {
		response.Error.Code = http.StatusInternalServerError
		return nil, &response.Error
	}

	return &response, nil
}

/**
 * UnauthorizedBody is a function that returns a pointer to a ZabbixRequest
 * It receives a string with the method and a ZabbixParams
 * It returns a pointer to an unauthorized ZabbixRequest
 * This function is used to make requests to the Zabbix API without a token
 */
func (z *ZabbixConnector) UnauthorizedBody(method string, params model.ZabbixParams) *model.ZabbixRequest {
	return &model.ZabbixRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		Id:      1,
	}
}

/**
 * Body is a function that returns a pointer to a ZabbixRequest
 * It receives a string with the method and a ZabbixParams
 * It returns a pointer to a ZabbixRequest
 */
func (z *ZabbixConnector) Body(method string, params model.ZabbixParams) *model.ZabbixRequest {
	if z.Token == "" {
		// die
		panic("Zabbix connection not initialized and required")
	}

	return &model.ZabbixRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		Auth:    z.Token,
		Id:      1,
	}
}

/**
 * Set the HTTP client for the ZabbixConnector
 */
func (c *ZabbixConnector) initializeClient() {
	c.client = req.C()
	c.client.SetCommonHeader("Content-Type", "application/json")
	c.client.SetBaseURL(c.Url)
	c.client.SetTimeout(5 * time.Second)
}

func (z *ZabbixConnector) GetVersion() (model.ZabbixVersion, *model.Error) {
	var version model.ZabbixVersion

	body := z.UnauthorizedBody("apiinfo.version", model.ZabbixParams{})

	response, err := z.Request(body)

	if err != nil {
		return version, err
	}

	versionString := response.Result.(string)[0:1]

	switch versionString {
	case "4":
		version = VERSION_40
	case "6":
		version = VERSION_60
	default:
		version = model.VERSION_UNKNOWN
	}

	if version == model.VERSION_UNKNOWN {
		return version, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Zabbix version not supported",
			Data:    versionString,
		}
	}

	return version, nil
}
