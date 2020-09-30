package aws_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/nabbar/golib/httpcli"

	"github.com/hashicorp/go-uuid"
	"github.com/nabbar/golib/aws"
	"github.com/nabbar/golib/aws/configCustom"
	"github.com/nabbar/golib/password"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	cli       aws.AWS
	cfg       aws.Config
	ctx       context.Context
	cnl       context.CancelFunc
	filename  = "./config.json"
	minioMode = false
)

/*
	Using https://onsi.github.io/ginkgo/
	Running with $> ginkgo -cover .
*/

func TestGolibAwsHelper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aws Helper Suite")
}

var _ = BeforeSuite(func() {
	var (
		err  error
		name string
		htp  *http.Client
	)

	ctx, cnl = context.WithCancel(context.Background())

	if err = loadConfig(); err != nil {
		var (
			uri = &url.URL{
				Scheme: "http",
				Host:   "localhost:" + strconv.Itoa(GetFreePort()),
			}

			accessKey = password.Generate(20)
			secretKey = password.Generate(64)
		)

		htp = httpcli.GetClient(uri.Hostname())
		Expect(htp).NotTo(BeNil())

		cfg = configCustom.NewConfig("", accessKey, secretKey, uri, "us-east-1")
		Expect(cfg).NotTo(BeNil())

		cfg.SetRegion("us-east-1")
		err = cfg.RegisterRegionAws(nil)
		Expect(err).NotTo(HaveOccurred())

		minioMode = true

		go LaunchMinio(uri.Host, accessKey, secretKey)

		for WaitMinio(uri.Host) {
			time.Sleep(10 * time.Second)
		}

		time.Sleep(10 * time.Second)
		println("Minio is waiting on : " + uri.Host)
	}

	cli, err = aws.New(ctx, cfg, htp)
	Expect(err).NotTo(HaveOccurred())
	Expect(cli).NotTo(BeNil())

	cli.ForcePathStyle(true)

	name, err = uuid.GenerateUUID()
	Expect(err).ToNot(HaveOccurred())
	Expect(name).ToNot(BeEmpty())
	cli.Config().SetBucketName(name)
})

var _ = AfterSuite(func() {
	cnl()
})

func loadConfig() error {
	var (
		cnfByt []byte
		err    error
	)

	if _, err = os.Stat(filename); err != nil {
		return err
	}

	if cnfByt, err = ioutil.ReadFile(filename); err != nil {
		return err
	}

	if cfg, err = configCustom.NewConfigJsonUnmashal(cnfByt); err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return nil
}

func BuildPolicy() string {
	return `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:Get*"],"Resource":["arn:aws:s3:::*/*"]}]}`
}

func BuildRole() string {
	return `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"sts:AssumeRole","Principal":{"Service":"replication"}}]}`
}

func GetFreePort() int {
	var (
		addr *net.TCPAddr
		lstn *net.TCPListener
		err  error
	)

	if addr, err = net.ResolveTCPAddr("tcp", "localhost:0"); err != nil {
		panic(err)
	}

	if lstn, err = net.ListenTCP("tcp", addr); err != nil {
		panic(err)
	}

	defer func() {
		_ = lstn.Close()
	}()

	return lstn.Addr().(*net.TCPAddr).Port
}

func GetTempFolder() string {
	if tmp, err := ioutil.TempDir("", "minio-data-*"); err != nil {
		panic(err)
	} else {
		if _, err = os.Stat(tmp); errors.Is(err, os.ErrNotExist) {
			if err = os.Mkdir(tmp, 0700); err != nil {
				panic(err)
			}
		} else if err != nil {
			panic(err)
		}

		return tmp
	}
}

func DelTempFolder(folder string) {
	if err := os.RemoveAll(folder); err != nil {
		panic(err)
	}
}

func LaunchMinio(host, accessKey, secretKey string) {
	os.Setenv("MINIO_ACCESS_KEY", accessKey)
	os.Setenv("MINIO_SECRET_KEY", secretKey)

	tmp := GetTempFolder()
	defer DelTempFolder(tmp)

	if _, minio, _, ok := runtime.Caller(0); ok {
		if err := exec.CommandContext(ctx, path.Join(path.Dir(minio), "minio"), "server", "--address", host, tmp).Run(); err != nil {
			panic(err)
		}
	} else {
		//nolint #goerr113
		panic(fmt.Errorf("minio execution file not found"))
	}

	//minio.Main([]string{"minio", "server", "--address", host, tmp})
}

func WaitMinio(host string) bool {
	conn, err := net.DialTimeout("tcp", host, 10*time.Second)

	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()

	return err == nil
}

func randContent(size int) *bytes.Buffer {
	p := make([]byte, size)

	_, err := rand.Read(p)

	if err != nil {
		panic(err)
	}

	return bytes.NewBuffer(p)
}
