// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

func (a *App) GetLMSToUse(consumerKey string) model.LMS {
	for _, lms := range a.Config().LTISettings.GetKnownLMSs() {
		if lms.GetOAuthConsumerKey() == consumerKey {
			return lms
		}
	}
	return nil
}

func (a *App) OnboardLMSUser(userId string, lms model.LMS, launchData map[string]string) *model.AppError {
	teamSlug := lms.GetTeam(launchData)
	team, err := a.GetTeamByName(teamSlug)
	if err != nil {
		mlog.Error(fmt.Sprintf("Team to be used: %s could not be found: %s", teamSlug, err.Error()))
		return model.NewAppError("OnboardLMSUser", "app.onboard_lms_user.team_not_found.app_error", nil, "", http.StatusInternalServerError)
	}

	if _, err := a.GetTeamMember(team.Id, userId); err != nil {
		// user is not a member of team. Adding team member
		if _, err := a.AddTeamMember(team.Id, userId); err != nil {
			mlog.Error(fmt.Sprintf("Error occurred while adding user %s to team %s: %s", userId, team.Id, err.Error()))
		}
	}

	a.createAndJoinChannels(team.Id, lms.GetPublicChannelsToJoin(launchData), model.CHANNEL_OPEN, userId)
	a.createAndJoinChannels(team.Id, lms.GetPrivateChannelsToJoin(launchData), model.CHANNEL_PRIVATE, userId)

	return nil
}

func (a *App) createAndJoinChannels(teamId string, channels map[string]string, channelType string, userId string) {
	for slug, displayName := range channels {
		channel, err := a.GetChannelByName(slug, teamId, true)
		if err != nil {
			// channel doesnt exist, create it
			channel = &model.Channel{
				TeamId:      teamId,
				Type:        channelType,
				Name:        slug,
				DisplayName: displayName,
			}

			channel, err = a.CreateChannel(channel, false)
			if err != nil {
				mlog.Error("Failed to create channel for LMS onboardin: " + err.Error())
			}
		}

		if _, err := a.AddChannelMember(userId, channel, "", "", false); err != nil {
			mlog.Error("Error occurred while adding user ID: " + userId + " to channel " + displayName + ": " + err.Error())
		}
	}
}
