package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	log "user-sync/logging"
	"user-sync/usersync"

	gapi "github.com/grafana/grafana-api-golang-client"
)

const BASE_URL_ENV_KEY string = "GRAFANA_SYNC_BASE_URL"
const ADMIN_USERNAME_ENV_KEY string = "GRAFANA_SYNC_ADMIN_USERNAME"
const ADMIN_PASSWORD_ENV_KEY string = "GRAFANA_SYNC_ADMIN_PASSWORD"
const ROOT_ORG_ID_ENV_KEY string = "GRAFANA_SYNC_ROOT_ORG_ID"

func main() {
	var baseUrlFlag = flag.String("url", "", "Base URL for the Grafana Instance")
	var adminUserNameFlag = flag.String("username", "", "Admin username for Grafana API")
	var adminPasswordFlag = flag.String("password", "", "Admin password for Grafana API")
	var rootOrgIdFlag = flag.String("rootorg", "1", "Root Grafana Org ID. Used as the base org to sync users from")
	flag.Parse()

	log.Init()

	baseUrl := getEnv(BASE_URL_ENV_KEY, DerefString(baseUrlFlag))
	adminUsername := getEnv(ADMIN_USERNAME_ENV_KEY, DerefString(adminUserNameFlag))
	adminPassword := getEnv(ADMIN_PASSWORD_ENV_KEY, DerefString(adminPasswordFlag))
	rootOrgId, _ := strconv.ParseInt(getEnv(ROOT_ORG_ID_ENV_KEY, DerefString(rootOrgIdFlag)), 10, 64)

	_, err := validateInputs(baseUrl, adminUsername, adminPassword, rootOrgId)

	if err != nil {
		log.Error().Println(err)
		return
	}

	var apiConfig = gapi.Config{
		BasicAuth: url.UserPassword(adminUsername, adminPassword),
		OrgID:     rootOrgId,
	}

	log.Info().Println("Starting Grafana User Sync Process...")

	client, err := gapi.New(baseUrl, apiConfig)

	if err != nil {
		log.Error().Println(err)
		return
	}

	success, err := usersync.PerformUserSync(usersync.UserSyncConfig{
		RootOrgId:           rootOrgId,
		ServerAdminUsername: adminUsername,
	}, client)

	if err != nil {
		log.Error().Println(err)
		return
	}

	if !success {
		log.Error().Println(errors.New("User sync process did not indicate an error but, it also did not succeed..."))
		return
	}

	log.Info().Println("Grafana User Sync Process completed successfully!")
	return

}

func validateInputs(url string, username string, password string, rootOrgId int64) (s bool, err error) {
	if url == "" {
		return false, errors.New(fmt.Sprintf("'url' is required. '--url http://grafana.local' or %s environment variable.", BASE_URL_ENV_KEY))
	}

	if username == "" {
		return false, errors.New(fmt.Sprintf("'username' is required. '--username admin' or %s environment variable.", ADMIN_USERNAME_ENV_KEY))
	}

	if password == "" {
		return false, errors.New(fmt.Sprintf("'password' is required. '--password 123Admin! or %s environment variable.'", ADMIN_PASSWORD_ENV_KEY))
	}

	if rootOrgId == 0 {
		return false, errors.New(fmt.Sprintf("'rootorg' is not a number. '--rootorg 1' or %s environment variable.", ROOT_ORG_ID_ENV_KEY))
	}

	return true, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func DerefString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}

func DerefInt64(i *int64) int64 {
	if i != nil {
		return *i
	}

	return 0
}
