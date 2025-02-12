package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"plugin"

	"github.com/npmaile/ScriptPortal/pkg/endpoints/httpHelpers"
	"github.com/npmaile/ScriptPortal/pkg/endpoints/scriptrunner"
	"github.com/npmaile/ScriptPortal/pkg/globals"
)

var port int
var pluginDirPath string

func init() {
	flag.IntVar(&port, "p", 80, "This is the port you wish to serve on")
	flag.StringVar(&globals.ScriptConfigPath, "c", "/var/lib/ScriptPortal/scriptConfig.json", "the path to your configuration json file")
	flag.StringVar(&pluginDirPath, "plug", "/var/opt/ScriptPortal/", "the path of your compiled .so plugin files")
	flag.StringVar(&globals.TemplatePath, "template", "/usr/share/ScriptPortal/templates/", "the path of the go http template files")
	flag.Parse()

}

func generateHandleFunc(targetFunc func(http.ResponseWriter, *http.Request, chan []byte)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		retChannel := make(chan []byte, 1)
		targetFunc(w, r, retChannel)
		var toAddToPage []byte
		//if the channel has been closed before a write, assume the plugin handlefunc took care of it. otherwise, put the response in the normal page boilerplate.
		for x := range retChannel {
			toAddToPage = x
		}
		if len(toAddToPage) > 0 {
			httpHelpers.WrapInBoilerPlate(w, string(toAddToPage))
		}
		return

	}
}

func main() {
	err := scriptrunner.Init(globals.ScriptConfigPath)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	plugDirInfo, _ := ioutil.ReadDir(pluginDirPath)
	httpHelpers.Navs = append(httpHelpers.Navs, httpHelpers.TopBarInfo{Name: "Scripts", Link: "/"})
	var plugins []string
	for _, x := range plugDirInfo {
		plugins = append(plugins, x.Name())
	}
	for _, pluginPath := range plugins {
		plug, err := plugin.Open(pluginDirPath + "/" + pluginPath)
		if err != nil {
			log.Fatalf("error loading Plugin: %s, %s\n", pluginDirPath, err.Error())
			return
		}
		handlefunc, err := plug.Lookup("HandleRequest")
		if err != nil {
			fmt.Printf("error loading symbol HandleRequest: %s, %s\n", pluginDirPath, err.Error())
		}
		NameFunc, err := plug.Lookup("Name")
		if err != nil {
			fmt.Printf("error loading symbol Name: %s, %s\n", pluginDirPath, err.Error())
			return
		}
		PathFunc, err := plug.Lookup("Path")
		if err != nil {
			fmt.Printf("error loading Symbol Path: %s, %s", pluginDirPath, err.Error())
		}
		httpHelpers.Navs = append(httpHelpers.Navs, httpHelpers.TopBarInfo{Name: NameFunc.(func() string)(), Link: PathFunc.(func() string)()})
		//Have fun parsing this one!
		http.HandleFunc(PathFunc.(func() string)(), generateHandleFunc(handlefunc.(func(http.ResponseWriter, *http.Request, chan []byte))))
	}
	http.HandleFunc("/", scriptrunner.ScriptIndex)
	http.HandleFunc("/script/", scriptrunner.ScriptMaster)
	http.HandleFunc("/status/", scriptrunner.HandleSocketRequest)
	http.HandleFunc("/search/", scriptrunner.ScriptSearch)
	fmt.Printf("Script Portal is running on port %d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Printf("it screwed up here: %s", err)
	}

}
