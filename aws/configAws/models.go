package configAws

import (
	"context"
	"fmt"
	"net"
	"net/url"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	libval "github.com/go-playground/validator/v10"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/httpcli"
)

type configModel struct {
	Region    string `mapstructure:"region" json:"region" yaml:"region" toml:"region" validate:"printascii,required"`
	AccessKey string `mapstructure:"accesskey" json:"accesskey" yaml:"accesskey" toml:"accesskey" validate:"printascii,required"`
	SecretKey string `mapstructure:"secretkey" json:"secretkey" yaml:"secretkey" toml:"secretkey" validate:"printascii,required"`
	Bucket    string `mapstructure:"bucket" json:"bucket" yaml:"bucket" toml:"bucket" validate:"printascii,omitempty"`
}

type awsModel struct {
	configModel
	retryer sdkaws.Retryer
}

func (c *awsModel) Validate() errors.Error {
	val := libval.New()
	err := val.Struct(c)

	if e, ok := err.(*libval.InvalidValidationError); ok {
		return ErrorConfigValidator.ErrorParent(e)
	}

	out := ErrorConfigValidator.Error(nil)

	for _, e := range err.(libval.ValidationErrors) {
		//nolint goerr113
		out.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Field(), e.ActualTag()))
	}

	if out.HasParent() {
		return out
	}

	return nil
}

func (c *awsModel) ResetRegionEndpoint() {
}

func (c *awsModel) RegisterRegionEndpoint(region string, endpoint *url.URL) errors.Error {
	return nil
}

func (c *awsModel) RegisterRegionAws(endpoint *url.URL) errors.Error {
	return nil
}

func (c *awsModel) SetRegion(region string) {
	c.Region = region
}

func (c *awsModel) GetRegion() string {
	return c.Region
}

func (c *awsModel) SetEndpoint(endpoint *url.URL) {
}

func (c awsModel) GetEndpoint() *url.URL {
	return nil
}

func (c *awsModel) ResolveEndpoint(service, region string) (sdkaws.Endpoint, error) {
	return sdkaws.Endpoint{}, ErrorEndpointInvalid.Error(nil)
}

func (c *awsModel) IsHTTPs() bool {
	return true
}

func (c *awsModel) SetRetryer(retryer sdkaws.Retryer) {
	c.retryer = retryer
}

func (c awsModel) Check(ctx context.Context) errors.Error {
	var (
		cfg *sdkaws.Config
		con net.Conn
		end sdkaws.Endpoint
		adr *url.URL
		err error
		e   errors.Error
	)

	if cfg, e = c.GetConfig(nil); e != nil {
		return e
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if end, err = cfg.EndpointResolver.ResolveEndpoint("s3", c.GetRegion()); err != nil {
		return ErrorEndpointInvalid.ErrorParent(err)
	}

	if adr, err = url.Parse(end.URL); err != nil {
		return ErrorEndpointInvalid.ErrorParent(err)
	}

	if _, err = cfg.Credentials.Retrieve(ctx); err != nil {
		return ErrorCredentialsInvalid.ErrorParent(err)
	}

	d := net.Dialer{
		Timeout:   httpcli.TIMEOUT_5_SEC,
		KeepAlive: httpcli.TIMEOUT_5_SEC,
	}

	con, err = d.DialContext(ctx, "tcp", adr.Host)

	defer func() {
		if con != nil {
			_ = con.Close()
		}
	}()

	if err != nil {
		return ErrorEndpointInvalid.ErrorParent(err)
	}

	return nil
}
