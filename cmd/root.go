package cmd

import (
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
var Version string = "__VERSION__"

type iflags struct {
	once *bool
}

var flags iflags

// initiations
var client = &http.Client{}

// variables
var ip string
var ipChanged bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "do-dyndns",
	Short:   "Dynamically set your subdomains IP addresses that utilize Digital Ocean nameservers.",
	Version: Version,
	PreRun: func(cmd *cobra.Command, args []string) {
		preRun(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) { run(cmd, args) },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		utils.Log.Fatal(err)
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
			results := make(chan bool, 1)
			go domainWorker(results)
			<-results
		}

		// break if running once
		if *flags.once {
			utils.Log.Infoln("Running once only.")
			break
		}

		// sleep for predefined time
		utils.Log.WithFields(log.Fields{"action": "HALT"}).Debugf("Wait for %ds\n", utils.Config.Interval)
		time.Sleep(time.Second * time.Duration(utils.Config.Interval))
	}

	utils.Log.Infoln("Finished...")

}

func domainWorker(end chan bool) {
	// create worker queue for domains
	numJobs := len(utils.Config.Domains)
	jobs := make(chan string, numJobs)
	results := make(chan string, numJobs)

	// create worker queue for subdomains
	var wg sync.WaitGroup

	// initialize variables
	var domainsProcessed []string
	var subdomainsProcessed []string

	// initiate workers
	for w := 0; w < numJobs; w++ {

		go func() {
			// create work group
			for domain := range jobs {

				// check if valid domain
				re := *regexp.MustCompile(`^[^.]*\.[^.]*$`)
				if !re.MatchString(domain) {
					utils.Log.WithFields(log.Fields{"component": "DOMAIN", "action": "SKIPPED"}).Errorln("Not a valid domain name:", domain)

					results <- domain
					return
				}

				utils.Log.WithFields(log.Fields{"component": "DOMAIN", "action": "STARTED"}).Debugln("Worker of domain:", domain)

				domainRecords, err := getDoDomainRecords(domain)

				if err != nil {
					utils.Log.WithFields(log.Fields{"component": "DO", "action": "CHECK"}).Errorln(err)

					results <- ""
					return
				}

				// match subdomains with domains
				var tempWG sync.WaitGroup
				for _, subdomain := range utils.Config.Subdomains {
					tempWG.Add(1)
					matched, _ := regexp.MatchString(fmt.Sprintf(`\.?%s$`, domain), subdomain)

					if matched {
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
									utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "CHANGE"}).Warningf(`Subdomain "%s" has different record of %s`, subdomain, record.Data)

									// send this to subdomain queue
									wg.Add(1)
									go subdomainWorker(&wg, domain, record.ID)
								}
							}
						}

						if !recordFound {
							utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "SKIPPED"}).Errorf(`Subdomain "%s" does not have a "A" record`, subdomain)

						} else {
							subdomainsProcessed = append(subdomainsProcessed, subdomain)

						}

					}

					tempWG.Done()
				}

				tempWG.Wait()

				utils.Log.WithFields(log.Fields{"component": "DOMAIN", "action": "FINISHED"}).Debugln("Worker of domain:", domain)

				domainsProcessed = append(domainsProcessed, domain)
				results <- domain
			}
		}()
	}

	// assign jobs
	for j := 0; j < numJobs; j++ {
		jobs <- utils.Config.Domains[j]
	}
	close(jobs)

	// wait for them to finish
	for a := 0; a < numJobs; a++ {
		<-results
	}

	// print out errors
	notProcessedDomains := getMissing(utils.Config.Domains, domainsProcessed)
	notProcessedSubdomains := getMissing(utils.Config.Subdomains, subdomainsProcessed)

	if len(notProcessedDomains) > 0 {
		utils.Log.WithFields(log.Fields{"component": "DOMAIN", "action": "FAILED"}).Errorf("Failed for domains: %s", strings.Join(notProcessedDomains, ", "))
	}

	if len(notProcessedSubdomains) > 0 {
		utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "FAILED"}).Errorf("Failed for subdomains: %s", strings.Join(notProcessedSubdomains, ", "))
	}

	// end run
	wg.Wait()
	end <- true
}

func subdomainWorker(wg *sync.WaitGroup, domain string, subdomainId int) {
	defer wg.Done()

	utils.Log.WithFields(log.Fields{"component": "SUBDOMAIN", "action": "STARTED"}).Debugln("Worker of subdomain:", domain)

}

func getIP() {
	utils.Log.WithFields(log.Fields{"component": "IP", "action": "START"}).Infoln("Querying for the API for IP address")

	// create request
	req, _ := http.NewRequest("GET", "https://api.ipify.org", nil)
	resp, err := client.Do(req)
	if err != nil {
		utils.Log.WithFields(log.Fields{"component": "IP", "action": "CHECK"}).Warnf("Can not create request to the API server. %s\n", err)

	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.Log.WithFields(log.Fields{"component": "IP", "action": "CHECK"}).Warnf("Can not connect to the API server. %s\n", err)
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

type doAPIResponse struct {
	DomainRecords doDomainRecords `json:"domain_records"`
	Links         struct {
	} `json:"links"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

type doDomainRecords []struct {
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

type doAPIErr struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func getDoDomainRecords(domain string) (doDomainRecords, error) {
	// create request
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.digitalocean.com/v2/domains/%s/records", domain), nil)
	req.Header.Add("User-Agent", "do-dyndns")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", utils.Config.Token))
	resp, err := client.Do(req)
	if err != nil {
		utils.Log.WithFields(log.Fields{"component": "DO", "action": "CHECK"}).Warnf("Can not create request to the API server. %s\n", err)

	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.Log.WithFields(log.Fields{"component": "IP", "action": "CHECK"}).Warnf("Can not connect to the API server. %s\n", err)

	}

	var apiErr doAPIErr
	json.Unmarshal(body, &apiErr)

	if apiErr.ID == "not_found" {
		return doDomainRecords{}, errors.New(fmt.Sprint("Records for the given domain can not be found: ", domain))
	}

	var value doAPIResponse
	json.Unmarshal(body, &value)

	// clean up records, only include A
	var domainRecords doDomainRecords
	for _, record := range value.DomainRecords {
		if record.Type == "A" {
			domainRecords = append(domainRecords, record)
		}
	}

	// check the length of A records
	if len(domainRecords) == 0 {
		return doDomainRecords{}, errors.New(fmt.Sprint("No A Records for given domain has been found: ", domain))
	}

	return domainRecords, nil
}

func getMissing(a, b []string) []string {
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
	rootCmd.PersistentFlags().StringVar(&utils.Cfg, "config", "", "config file ({.,/etc/do-dyndns/,$HOME}/.do-dyndns.yml)")
	rootCmd.PersistentFlags().BoolVar(&utils.LogLevelVerbose, "verbose", false, "Enable verbose logging.")

	// initialize
	cobra.OnInitialize(utils.InitiateLogger, utils.InitConfig)

	// command flags
	flags = iflags{once: rootCmd.Flags().BoolP("once", "o", false, "Run the command only once.")}
}
