package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	utils "github.com/cenk1cenk2/do-dyndns/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Version get current version of application
var Version string = "1.0.0"

type iflags struct {
	once *bool
}

var flags iflags

// variables
var ip string
var ipChanged bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "do-dyndns",
	Short:   "Dynamically set your subdomains IP addresses that utilize Digital Ocean nameservers.",
	Example: `Please visit url for readme "https://github.com/cenk1cenk2/do-dyndns/blob/master/README.md"`,
	Version: Version,
	PreRun:  func(cmd *cobra.Command, args []string) { preRun(cmd, args) },
	Run:     func(cmd *cobra.Command, args []string) { run(cmd, args) },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		utils.Log.Fatalln(err)
		os.Exit(1)
	}
}

func preRun(cmd *cobra.Command, args []string) {
	// Load configuration
	utils.LoadConfig()
}

func run(cmd *cobra.Command, args []string) {
	// create a new set of domain and subdomain workers
	for {
		// query for ip address
		getIP()

		// if ip address changed initiate the process
		if ipChanged {

			domainWorker()

			ipChanged = false
		} else {
			utils.Log.WithFields(log.Fields{"action": "HALT"}).Debugf("Ip address has not changed.")
		}

		// break if running once
		if *flags.once {
			utils.Log.Infoln("Running once only.")
			break
		}

		// sleep for predefined time
		utils.Log.WithFields(log.Fields{"action": "HALT"}).Infof("Wait for %ds", utils.Config.Interval)
		time.Sleep(time.Second * time.Duration(utils.Config.Interval))
	}

	utils.Log.Infoln("Finished...")

}

func domainWorker() {
	var wg sync.WaitGroup

	// initialize variables
	var domainsProcessed []string
	var domainsProcessedMutex = &sync.Mutex{}
	var subdomainsProcessed []string
	var subdomainsProcessedMutex = &sync.Mutex{}

	// create work group
	for _, domain := range utils.Config.Domains {
		wg.Add(1)

		go func(domain string) {
			defer wg.Done()

			// check if valid domain
			re := *regexp.MustCompile(`^[^.]*\.[^.]*$`)
			if !re.MatchString(domain) {
				utils.Log.WithFields(log.Fields{"component": "DOMAIN", "action": "SKIPPED"}).Errorln("Not a valid domain name:", domain)

				return
			}

			utils.Log.WithFields(log.Fields{"component": "DOMAIN", "action": "STARTED"}).Debugln("Worker of domain:", domain)

			domainRecords, err := getDoDomainRecords(domain)

			if err != nil {
				utils.Log.WithFields(log.Fields{"component": "DO", "action": "CHECK"}).Errorln(err)

				return
			}

			// create subdomain wg
			var subdomainWG sync.WaitGroup

			// match subdomains with domains
			for _, subdomain := range utils.Config.Subdomains {
				matched, _ := regexp.MatchString(fmt.Sprintf(`\.?%s$`, domain), subdomain)

				if matched {
					subdomainWG.Add(1)

					go subdomainWorker(&subdomainWG, domain, subdomain, domainRecords, &subdomainsProcessed, subdomainsProcessedMutex)
				}

			}

			utils.Log.WithFields(log.Fields{"component": "DOMAIN", "action": "FINISHED"}).Debugln("Worker of domain:", domain)

			subdomainWG.Wait()

			domainsProcessedMutex.Lock()
			domainsProcessed = append(domainsProcessed, domain)
			domainsProcessedMutex.Unlock()
		}(domain)
	}

	wg.Wait()
	// print out errors
	notProcessedDomains := getMissingSlice(utils.Config.Domains, domainsProcessed)
	notProcessedSubdomains := getMissingSlice(utils.Config.Subdomains, subdomainsProcessed)

	if len(notProcessedDomains) > 0 {
		utils.Log.WithFields(log.Fields{"component": "DOMAIN", "action": "FAILED"}).Errorf("Failed for domains: %s", strings.Join(notProcessedDomains, ", "))
	}

	if len(notProcessedSubdomains) > 0 {
		utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "FAILED"}).Errorf("Failed for subdomains: %s", strings.Join(notProcessedSubdomains, ", "))
	}

}

func subdomainWorker(wg *sync.WaitGroup, domain string, subdomain string, domainRecords []iDoDomainRecordsAPI, subdomainsProcessed *[]string, subdomainsProcessedMutex *sync.Mutex) {
	defer wg.Done()

	utils.Log.WithField("component", "DOMAIN").Debugf("Matched domain %s with subdomain %s\n", domain, subdomain)

	var parsedSubdomain string

	// strip subdomain to bare
	if subdomain == domain {
		// subdomain is the root domain
		parsedSubdomain = "@"

	} else {
		// subdomain is a real subdomain
		re := *regexp.MustCompile(`^(.*)\.[^.]*\.[^.]*$`)
		parsedSubdomain = re.FindStringSubmatch(subdomain)[1]

	}

	// the case where you want to change the base domain name directly
	var recordFound bool
	for _, record := range domainRecords {
		if record.Name == parsedSubdomain {
			recordFound = true

			if record.Data != ip {
				utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "STARTED"}).Debugln("Worker of subdomain:", domain)
				utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "CHANGE"}).Warningf(`Subdomain "%s" has different record of %s`, subdomain, record.Data)

				res, err := setDoDomainRecords(domain, subdomain, record.ID)

				if err != nil {
					utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "FAILED"}).Errorln(err)

				} else {
					utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "SUCCESS"}).Infoln(res)

				}

			} else {
				utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "SKIPPED"}).Debugln("No changes required for subodmain:", subdomain)

			}
		}
	}

	if !recordFound {
		utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "SKIPPED"}).Errorf(`Subdomain "%s" does not have a "A" record`, subdomain)

	} else {
		subdomainsProcessedMutex.Lock()
		*subdomainsProcessed = append(*subdomainsProcessed, subdomain)
		subdomainsProcessedMutex.Unlock()

	}
}

func getIP() {
	utils.Log.WithFields(log.Fields{"component": "IP", "action": "START"}).Infoln("Fetching the IP address from the API")

	body, err := createAPIRequest("GET",
		"https://api.ipify.org",
		"",
	)

	if err != nil {
		utils.Log.WithFields(log.Fields{"component": "IP", "action": "CHECK"}).Warnln(err)

		ipChanged = false
	} else {
		// check ip address
		query := string(body)
		utils.Log.WithFields(log.Fields{"component": "IP", "action": "CHECK"}).Infoln("Current IP address is:", query)

		if query != ip {
			if ip != "" {
				utils.Log.WithFields(log.Fields{"component": "IP", "action": "CHANGE"}).Warningf("IP address has been changed from %s to %s\n", ip, query)
			}

			ip = query
			ipChanged = true

		} else {
			ipChanged = false

		}

	}

}

type iGetDoDomainRecordsAPIRes struct {
	DomainRecords []iDoDomainRecordsAPI `json:"domain_records"`
	Links         struct {
	} `json:"links"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

type iDoAPIErr struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type iSetDoDomainRecordsAPIRes struct {
	DomainRecord iDoDomainRecordsAPI `json:"domain_record"`
}

type iDoDomainRecordsAPI struct {
	ID       int         `json:"id"`
	Type     string      `json:"type"`
	Name     string      `json:"name"`
	Data     string      `json:"data"`
	Priority interface{} `json:"priority"`
	Port     interface{} `json:"port"`
	TTL      int         `json:"ttl"`
	Weight   interface{} `json:"weight"`
	Flags    interface{} `json:"flags"`
	Tag      interface{} `json:"tag"`
}

func getDoDomainRecords(domain string) ([]iDoDomainRecordsAPI, error) {
	body, err := createAPIRequest("GET",
		fmt.Sprintf("https://api.digitalocean.com/v2/domains/%s/records", domain),
		"",
		iRequestHeaders{Key: "content-type", Value: "application/json"},
		iRequestHeaders{Key: "Authorization", Value: fmt.Sprintf("Bearer %s", utils.Config.Token)},
	)

	if err != nil {
		return []iDoDomainRecordsAPI{}, err
	}

	var apiErr iDoAPIErr
	json.Unmarshal(body, &apiErr)

	if apiErr.ID == "not_found" {
		return []iDoDomainRecordsAPI{}, errors.New(fmt.Sprint("Records for the given domain can not be found: ", domain))
	}

	if apiErr.ID == "Unauthorized" {
		return []iDoDomainRecordsAPI{}, errors.New(fmt.Sprint("Token does not seem to be valid for domain: ", domain))
	}

	var value iGetDoDomainRecordsAPIRes
	json.Unmarshal(body, &value)

	// clean up records, only include A
	var domainRecords []iDoDomainRecordsAPI
	for _, record := range value.DomainRecords {
		if record.Type == "A" {
			domainRecords = append(domainRecords, record)
		}
	}

	fmt.Println(string(body))

	// check the length of A records
	if len(domainRecords) == 0 {
		return []iDoDomainRecordsAPI{}, errors.New(fmt.Sprint("No A Records for given domain has been found: ", domain))
	}

	return domainRecords, nil
}

func setDoDomainRecords(domain string, subdomain string, subdomainID int) (string, error) {
	body, err := createAPIRequest("PUT",
		fmt.Sprintf("https://api.digitalocean.com/v2/domains/%s/records/%d", domain, subdomainID),
		fmt.Sprintf(`{"data":"%s"}`, ip),
		iRequestHeaders{Key: "content-type", Value: "application/json"},
		iRequestHeaders{Key: "Authorization", Value: fmt.Sprintf("Bearer %s", utils.Config.Token)},
	)

	if err != nil {
		return "", err
	}

	var apiErr iDoAPIErr
	json.Unmarshal(body, &apiErr)

	if apiErr.ID == "not_found" {
		return "", errors.New(fmt.Sprint("Digital Ocean API rejected the record for subdomain:", subdomain))
	}

	if apiErr.ID == "Unauthorized" {
		return "", errors.New(fmt.Sprint("Token does not seem to be valid for subdomain: ", subdomain))
	}

	var value iSetDoDomainRecordsAPIRes
	json.Unmarshal(body, &value)

	var record = value.DomainRecord
	if record.Data != ip {
		return "", errors.New(fmt.Sprint("Digital Ocean API failed to set the record for subdomain:", subdomain))
	}

	return fmt.Sprintln("Changed DNS record for subdomain:", subdomain), nil
}

type iRequestHeaders struct {
	Value string
	Key   string
}

func createAPIRequest(method string, url string, data string, headers ...iRequestHeaders) ([]byte, error) {
	// initiations
	var client = &http.Client{}

	// create request
	req, _ := http.NewRequest(method, url, bytes.NewBuffer([]byte(data)))
	req.Header.Add("User-Agent", "do-dyndns")

	for _, header := range headers {
		req.Header.Add(header.Key, header.Value)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, errors.New(fmt.Sprint("Can not connect to the API server.", err))
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.New(fmt.Sprint("Can not decode the API response.", err))
	}

	return body, nil
}

func getMissingSlice(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func init() {
	fmt.Println("|d|o|-|d|y|n|d|n|s|", fmt.Sprintf("v%s", Version))

	// persistent flags
	rootCmd.PersistentFlags().StringVar(&utils.Cfg, "config", "", "config file ({.,/etc/do-dyndns,~/.config/do-dyndns}/.do-dyndns.yml)")
	rootCmd.PersistentFlags().BoolVar(&utils.LogLevelVerbose, "verbose", false, "Enable verbose logging.")

	// initialize
	cobra.OnInitialize(utils.InitiateLogger, utils.InitConfig)

	// command flags
	flags = iflags{once: rootCmd.Flags().BoolP("once", "o", false, "Run the command only once.")}
}
