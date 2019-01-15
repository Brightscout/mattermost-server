package web

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/utils"
)

func (w *Web) InitLti() {
	w.MainRouter.Handle("/login/lti", w.NewHandler(loginWithLti)).Methods("POST")
}

func loginWithLti(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.Config().LTISettings.Enable {
		c.Err =  model.NewAppError("loginWithLti", "api.lti.disabled.app_error", nil, "", http.StatusNotImplemented)
		return
	}

	// Validate request
	lmss := c.App.Config().LTISettings.LMSs
	ltiConsumerKey := r.FormValue("oauth_consumer_key")
	var ltiConsumerSecret string

	for _, val := range lmss {
		if lms, ok := val.(model.EdxLMSSettings); ok {
			if lms.OAuth.ConsumerKey == ltiConsumerKey {
				ltiConsumerSecret = lms.OAuth.ConsumerSecret
				break
			}
		}
	}

	p := utils.NewProvider(ltiConsumerSecret, fmt.Sprintf("%s%s", c.GetSiteURLHeader(), c.Path))
	p.ConsumerKey = ltiConsumerKey
	if ok, err := p.IsValid(r); err != nil || ok == false {
		// TODO: update this based on how we handle request validation error
		mlog.Error("Invalid LTI request: " + err.Error())
		c.Err =  model.NewAppError("loginWithLti", "api.lti.validate.app_error", nil, "", http.StatusNotImplemented)
		return
	}

	http.Redirect(w, r, c.GetSiteURLHeader(), http.StatusFound)
}
