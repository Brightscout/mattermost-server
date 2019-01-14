package web

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/app"
	"github.com/mattermost/mattermost-server/mlog"
)

func (w *Web) InitLti() {

	mlog.Info("Initializing web LTI")
	w.MainRouter.Handle("/login/lti", w.NewHandler(loginWithLti)).Methods("POST")
}

func loginWithLti(c *Context, w http.ResponseWriter, r *http.Request) {

	// Validate request
	lmss := c.App.Config().LTISettings.LMSs
	ltiConsumerKey := r.FormValue("oauth_consumer_key")
	var ltiConsumerSecret string

	for _, lms := range lmss {
		if lms.OAuth.ConsumerKey == ltiConsumerKey {
			ltiConsumerSecret = lms.OAuth.ConsumerSecret
			break
		}
	}

	p := app.NewProvider(ltiConsumerSecret, fmt.Sprintf("%s%s", c.GetSiteURLHeader(), c.Path))
	p.ConsumerKey = ltiConsumerKey
	if ok, err := p.IsValid(r); err != nil || ok == false {
		fmt.Fprintf(w, "Invalid request...")
		mlog.Error("Invalid request: " + err.Error())
		return
	}

	http.Redirect(w, r, c.GetSiteURLHeader(), http.StatusFound)
}
