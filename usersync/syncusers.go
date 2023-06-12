package usersync

import (
	"errors"
	"strings"
	log "user-sync/logging"

	gapi "github.com/grafana/grafana-api-golang-client"
)

type UserSyncConfig struct {
	RootOrgId           int64
	ServerAdminUsername string
}

func PerformUserSync(config UserSyncConfig, client *gapi.Client) (sucess bool, err error) {
	success, err := performUserSyncProcess(config, client)

	if err != nil {
		log.Error().Println(err)
	}

	return success, err
}

func performUserSyncProcess(config UserSyncConfig, client *gapi.Client) (success bool, err error) {
	orgs, err := getAllNonRootOrgs(config.RootOrgId, client)

	if err != nil {
		log.Error().Println(err)
		return false, err
	}

	if len(orgs) <= 0 {
		noOrgs := errors.New("No additional Grafana Orgs were found...")
		log.Error().Println(noOrgs)
		return false, noOrgs
	}

	logFoundOrgs(orgs)

	users, err := getAllUsersInRootOrg(config, client)

	if err != nil {
		log.Error().Println(err)
		return false, err
	}

	if len(users) <= 0 {
		noUsers := errors.New("No Users were found in the root org...")
		log.Error().Println(noUsers)
		return false, noUsers
	}

	logNumberUsersFound(len(users))

	for _, u := range users {
		addUserToOrgs(u, orgs, config, client)
	}

	return true, nil
}

func addUserToOrgs(user gapi.OrgUser, orgs []gapi.Org, config UserSyncConfig, client *gapi.Client) {
	for _, o := range orgs {
		success, err := addUserToOrg(user, o, config, client)

		if err != nil {
			if strings.Contains(err.Error(), "status: 409") {
				log.Warn().Printf("%s already exists in %s", user.Email, o.Name)
			} else {
				log.Error().Println(err)
			}
		}

		if !success {
			log.Warn().Printf("%s was not able to be synced to %s", user.Email, o.Name)
		} else {
			log.Info().Printf("%s was synced to %s", user.Email, o.Name)
		}
	}

}

func addUserToOrg(user gapi.OrgUser, org gapi.Org, config UserSyncConfig, client *gapi.Client) (success bool, e error) {
	log.Info().Printf("Attempting to add %s to the %s org", user.Email, org.Name)
	err := client.AddOrgUser(org.ID, user.Email, user.Role)

	return err == nil, err
}

func getAllUsersInRootOrg(config UserSyncConfig, client *gapi.Client) (users []gapi.OrgUser, err error) {
	resp, err := client.OrgUsers(config.RootOrgId)

	if err != nil {
		return nil, err
	}

	var nonAdminUsers = filter[gapi.OrgUser](resp, func(u gapi.OrgUser) bool {
		return u.Login != config.ServerAdminUsername
	})

	return nonAdminUsers, err
}

func getAllNonRootOrgs(rootOrg int64, client *gapi.Client) (orgs []gapi.Org, err error) {
	resp, err := client.Orgs()

	if err != nil {
		return nil, err
	}

	var nonRootOrgs = filter[gapi.Org](resp, func(o gapi.Org) bool {
		return o.ID != rootOrg
	})

	return nonRootOrgs, nil
}

func logFoundOrgs(orgs []gapi.Org) {
	names := selectValue[gapi.Org, string](orgs, func(o gapi.Org) string {
		return o.Name
	})

	msg := strings.Join(names, ", ")

	log.Info().Printf("Found the following Orgs: %s", msg)
}

func logNumberUsersFound(count int) {
	log.Info().Printf("Found %d users in the root org.", count)
}

func selectValue[T any, R any](ss []T, sel func(T) R) (ret []R) {
	for _, v := range ss {
		ret = append(ret, sel(v))
	}
	return
}

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
