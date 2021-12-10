package scriptrunner

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/npmaile/ScriptPortal/endpoints/httpHelpers"
	"github.com/npmaile/ScriptPortal/globals"
)

//ScriptIndex generates an index for scripts installed and properly configured in scriptConfig.json
func ScriptIndex(w http.ResponseWriter, r *http.Request) {
	scripts := configuredScripts
	templhtml, err := ioutil.ReadFile(globals.TemplatePath + "scriptIndex.html")
	if err != nil {
		fmt.Println(err)
	}
	httpHelpers.AddContentToPage(w, string(templhtml), scripts)
}
