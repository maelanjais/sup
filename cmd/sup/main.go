package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	hcplugin "github.com/hashicorp/go-plugin"
	"github.com/mikkeloscar/sshconfig"
	"github.com/pkg/errors"
	"github.com/pressly/sup"

	// Importer le package shared de ton plugin
	"github.com/maelanjais/sup-hcl2-plugin/shared"
)

var (
	supfile     string
	envVars     flagStringSlice
	sshConfig   string
	onlyHosts   string
	exceptHosts string

	debug         bool
	disablePrefix bool

	showVersion bool
	showHelp    bool

	// NOUVEAU: chemin vers le binaire plugin HCL2
	hcl2ParserPath string

	ErrUsage            = errors.New("Usage: sup [OPTIONS] NETWORK COMMAND [...]\n       sup [ --help | -v | --version ]")
	ErrUnknownNetwork   = errors.New("Unknown network")
	ErrNetworkNoHosts   = errors.New("No hosts defined for a given network")
	ErrCmd              = errors.New("Unknown command/target")
	ErrTargetNoCommands = errors.New("No commands defined for a given target")
	ErrConfigFile       = errors.New("Unknown ssh_config file")
)

type flagStringSlice []string

func (f *flagStringSlice) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *flagStringSlice) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func init() {
	flag.StringVar(&supfile, "f", "", "Custom path to ./Supfile[.yml|.hcl]")
	flag.Var(&envVars, "e", "Set environment variables")
	flag.Var(&envVars, "env", "Set environment variables")
	flag.StringVar(&sshConfig, "sshconfig", "", "Read SSH Config file, ie. ~/.ssh/config file")
	flag.StringVar(&onlyHosts, "only", "", "Filter hosts using regexp")
	flag.StringVar(&exceptHosts, "except", "", "Filter out hosts using regexp")

	flag.BoolVar(&debug, "D", false, "Enable debug mode")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&disablePrefix, "disable-prefix", false, "Disable hostname prefix")

	flag.BoolVar(&showVersion, "v", false, "Print version")
	flag.BoolVar(&showVersion, "version", false, "Print version")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	// NOUVEAU: flag pour le chemin du plugin
	flag.StringVar(&hcl2ParserPath, "parser", "", "Path to HCL2 parser plugin binary")
}

// NOUVEAU: parseHCLViaPlugin lance le plugin et convertit HCL → YAML bytes
func parseHCLViaPlugin(hclFilePath string) ([]byte, error) {
	// Trouver le binaire plugin
	parserBin := hcl2ParserPath
	if parserBin == "" {
		// Chercher dans le PATH par défaut
		var err error
		parserBin, err = exec.LookPath("sup-hcl2-parser")
		if err != nil {
			return nil, fmt.Errorf("HCL2 parser plugin not found. Install it or use --parser flag: %v", err)
		}
	}

	// Lancer le plugin via go-plugin
	client := hcplugin.NewClient(&hcplugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		Cmd:             exec.Command(parserBin),
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("plugin connection failed: %v", err)
	}

	raw, err := rpcClient.Dispense("config_parser")
	if err != nil {
		return nil, fmt.Errorf("plugin dispense failed: %v", err)
	}

	parser := raw.(shared.ConfigParser)
	return parser.ParseFile(hclFilePath)
}

func networkUsage(conf *sup.Supfile) {
	w := &tabwriter.Writer{}
	w.Init(os.Stderr, 4, 4, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintln(w, "Networks:\t")
	for _, name := range conf.Networks.Names {
		fmt.Fprintf(w, "- %v\n", name)
		network, _ := conf.Networks.Get(name)
		for _, host := range network.Hosts {
			fmt.Fprintf(w, "\t- %v\n", host)
		}
	}
	fmt.Fprintln(w)
}

func cmdUsage(conf *sup.Supfile) {
	w := &tabwriter.Writer{}
	w.Init(os.Stderr, 4, 4, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintln(w, "Targets:\t")
	for _, name := range conf.Targets.Names {
		cmds, _ := conf.Targets.Get(name)
		fmt.Fprintf(w, "- %v\t%v\n", name, strings.Join(cmds, " "))
	}
	fmt.Fprintln(w, "\t")
	fmt.Fprintln(w, "Commands:\t")
	for _, name := range conf.Commands.Names {
		cmd, _ := conf.Commands.Get(name)
		fmt.Fprintf(w, "- %v\t%v\n", name, cmd.Desc)
	}
	fmt.Fprintln(w)
}

func parseArgs(conf *sup.Supfile) (*sup.Network, []*sup.Command, error) {
	var commands []*sup.Command
	args := flag.Args()
	if len(args) < 1 {
		networkUsage(conf)
		return nil, nil, ErrUsage
	}
	network, ok := conf.Networks.Get(args[0])
	if !ok {
		networkUsage(conf)
		return nil, nil, ErrUnknownNetwork
	}
	for _, env := range envVars {
		if len(env) == 0 {
			continue
		}
		i := strings.Index(env, "=")
		if i < 0 {
			if len(env) > 0 {
				network.Env.Set(env, "")
			}
			continue
		}
		network.Env.Set(env[:i], env[i+1:])
	}
	hosts, err := network.ParseInventory()
	if err != nil {
		return nil, nil, err
	}
	network.Hosts = append(network.Hosts, hosts...)
	if len(network.Hosts) == 0 {
		networkUsage(conf)
		return nil, nil, ErrNetworkNoHosts
	}
	if len(args) < 2 {
		cmdUsage(conf)
		return nil, nil, ErrUsage
	}
	if network.Env == nil {
		network.Env = make(sup.EnvList, 0)
	}
	network.Env.Set("SUP_NETWORK", args[0])
	network.Env.Set("SUP_TIME", time.Now().UTC().Format(time.RFC3339))
	if os.Getenv("SUP_TIME") != "" {
		network.Env.Set("SUP_TIME", os.Getenv("SUP_TIME"))
	}
	if os.Getenv("SUP_USER") != "" {
		network.Env.Set("SUP_USER", os.Getenv("SUP_USER"))
	} else {
		network.Env.Set("SUP_USER", os.Getenv("USER"))
	}
	for _, cmd := range args[1:] {
		target, isTarget := conf.Targets.Get(cmd)
		if isTarget {
			for _, cmd := range target {
				command, isCommand := conf.Commands.Get(cmd)
				if !isCommand {
					cmdUsage(conf)
					return nil, nil, fmt.Errorf("%v: %v", ErrCmd, cmd)
				}
				command.Name = cmd
				commands = append(commands, &command)
			}
		}
		command, isCommand := conf.Commands.Get(cmd)
		if isCommand {
			command.Name = cmd
			commands = append(commands, &command)
		}
		if !isTarget && !isCommand {
			cmdUsage(conf)
			return nil, nil, fmt.Errorf("%v: %v", ErrCmd, cmd)
		}
	}
	return &network, commands, nil
}

func resolvePath(path string) string {
	if path == "" {
		return ""
	}
	if path[:2] == "~/" {
		usr, err := user.Current()
		if err == nil {
			path = filepath.Join(usr.HomeDir, path[2:])
		}
	}
	return path
}

func main() {
	flag.Parse()

	if showHelp {
		fmt.Fprintln(os.Stderr, ErrUsage, "\n\nOptions:")
		flag.PrintDefaults()
		return
	}
	if showVersion {
		fmt.Fprintln(os.Stderr, sup.VERSION)
		return
	}

	// ===== SECTION MODIFIÉE: Support HCL2 via plugin =====
	if supfile == "" {
		supfile = "./Supfile"
	}

	var data []byte
	var err error
	resolvedPath := resolvePath(supfile)

	// Détecter l'extension .hcl
	if strings.HasSuffix(resolvedPath, ".hcl") {
		// Utiliser le plugin HCL2
		data, err = parseHCLViaPlugin(resolvedPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "HCL2 plugin error:", err)
			os.Exit(1)
		}
	} else {
		// Fallback YAML classique
		data, err = ioutil.ReadFile(resolvedPath)
		if err != nil {
			firstErr := err
			data, err = ioutil.ReadFile("./Supfile.yml")
			if err != nil {
				// Tenter aussi .hcl en dernier recours
				data, err = parseHCLViaPlugin("./Supfile.hcl")
				if err != nil {
					fmt.Fprintln(os.Stderr, firstErr)
					os.Exit(1)
				}
			}
		}
	}
	// ===== FIN SECTION MODIFIÉE =====

	conf, err := sup.NewSupfile(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	network, commands, err := parseArgs(conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if onlyHosts != "" {
		expr, err := regexp.CompilePOSIX(onlyHosts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		var hosts []string
		for _, host := range network.Hosts {
			if expr.MatchString(host) {
				hosts = append(hosts, host)
			}
		}
		if len(hosts) == 0 {
			fmt.Fprintln(os.Stderr, fmt.Errorf("no hosts match --only '%v' regexp", onlyHosts))
			os.Exit(1)
		}
		network.Hosts = hosts
	}

	if exceptHosts != "" {
		expr, err := regexp.CompilePOSIX(exceptHosts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		var hosts []string
		for _, host := range network.Hosts {
			if !expr.MatchString(host) {
				hosts = append(hosts, host)
			}
		}
		if len(hosts) == 0 {
			fmt.Fprintln(os.Stderr, fmt.Errorf("no hosts left after --except '%v' regexp", exceptHosts))
			os.Exit(1)
		}
		network.Hosts = hosts
	}

	if sshConfig != "" {
		confHosts, err := sshconfig.ParseSSHConfig(resolvePath(sshConfig))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		confMap := map[string]*sshconfig.SSHHost{}
		for _, conf := range confHosts {
			for _, host := range conf.Host {
				confMap[host] = conf
			}
		}
		for _, host := range network.Hosts {
			conf, found := confMap[host]
			if found {
				network.User = conf.User
				network.IdentityFile = resolvePath(conf.IdentityFile)
				network.Hosts = []string{fmt.Sprintf("%s:%d", conf.HostName, conf.Port)}
			}
		}
	}

	var vars sup.EnvList
	for _, val := range append(conf.Env, network.Env...) {
		vars.Set(val.Key, val.Value)
	}
	if err := vars.ResolveValues(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var cliVars sup.EnvList
	for _, env := range envVars {
		if len(env) == 0 {
			continue
		}
		i := strings.Index(env, "=")
		if i < 0 {
			if len(env) > 0 {
				vars.Set(env, "")
			}
			continue
		}
		vars.Set(env[:i], env[i+1:])
		cliVars.Set(env[:i], env[i+1:])
	}

	supEnv := ""
	for _, v := range cliVars {
		supEnv += fmt.Sprintf(" -e %v=%q", v.Key, v.Value)
	}
	vars.Set("SUP_ENV", strings.TrimSpace(supEnv))

	app, err := sup.New(conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	app.Debug(debug)
	app.Prefix(!disablePrefix)

	err = app.Run(network, vars, commands...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
