package plugin_installer

import (
	"flag"
	"os"

	"github.com/spf13/pflag"
	utilflag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/plugins"
)

func BindFlags(flags *pflag.FlagSet, plugin bool) clientcmd.ClientConfig {
	flags.AddGoFlagSet(flag.CommandLine)
	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	clientConfig := util.DefaultClientConfig(flags)
	if plugin {
		loadFromEnv(flags, "kubeconfig")
		loadFromEnv(flags, clientcmd.FlagClusterName)
		loadFromEnv(flags, clientcmd.FlagAuthInfoName)
		loadFromEnv(flags, clientcmd.FlagContext)
		loadFromEnv(flags, clientcmd.FlagNamespace)
		loadFromEnv(flags, clientcmd.FlagAPIServer)
		loadFromEnv(flags, clientcmd.FlagInsecure)
		loadFromEnv(flags, clientcmd.FlagCertFile)
		loadFromEnv(flags, clientcmd.FlagKeyFile)
		loadFromEnv(flags, clientcmd.FlagCAFile)
		loadFromEnv(flags, clientcmd.FlagBearerToken)
		loadFromEnv(flags, clientcmd.FlagImpersonate)
		loadFromEnv(flags, clientcmd.FlagImpersonateGroup)
		loadFromEnv(flags, clientcmd.FlagUsername)
		loadFromEnv(flags, clientcmd.FlagPassword)
		loadFromEnv(flags, clientcmd.FlagTimeout)

		loadFromEnv(flags, "alsologtostderr")
		loadFromEnv(flags, "log-backtrace-at")
		loadFromEnv(flags, "log-dir")
		loadFromEnv(flags, "logtostderr")
		loadFromEnv(flags, "stderrthreshold")
		loadFromEnv(flags, "v")
		loadFromEnv(flags, "vmodule")
	}
	return clientConfig
}

func loadFromEnv(flags *pflag.FlagSet, name string) {
	v, found := os.LookupEnv(plugins.FlagToEnvName(name, "KUBECTL_PLUGINS_GLOBAL_FLAG_"))
	if found && (name != clientcmd.FlagImpersonateGroup || v != "[]") {
		flags.Set(name, v)
	}
}
