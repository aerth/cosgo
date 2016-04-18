package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/aerth/seconf"
)

func backwardsComp() {

	// For backwards compatibility
	if os.Getenv("CASGO_API_KEY") != "" && os.Getenv("COSGO_API_KEY") == "" {
		os.Setenv("COSGO_API_KEY", os.Getenv("CASGO_API_KEY"))
		log.Println("Please use COSGO_API_KEY instead of depreciated CASGO_API_KEY")
	}
	if os.Getenv("CASGO_DESTINATION") != "" && os.Getenv("COSGO_DESTINATION") == "" {
		os.Setenv("COSGO_DESTINATION", os.Getenv("CASGO_DESTINATION"))
		log.Println("Please use COSGO_DESTINATION instead of depreciated CASGO_DESTINATION")
	}

}

func quickSelfTest() (err error) {
	// If not using config, and mailbox is not requested, require a SMTP API key. Otherwise, go for mailbox mode.
	if !*config {
		if mailbox == false {
			switch *mailmode {
			case smtpmandrill:
				mandrillKey = os.Getenv("MANDRILL_KEY")
				if mandrillKey == "" {
					return errors.New("Fatal: environmental variable `MANDRILL_KEY` is Crucial.\n\n\t\tHint: export MANDRILL_KEY=123456789")

				}
				if cosgoDestination == "" {
					cosgoDestination = os.Getenv("COSGO_DESTINATION")
					if cosgoDestination == "" {
						return errors.New("Fatal: environmental variable `COSGO_DESTINATION` is Crucial.\n\n\t\tHint: export COSGO_DESTINATION=you@yours.com")
					}
				}
			case smtpsendgrid:
				sendgridKey = os.Getenv("SENDGRID_KEY")
				if sendgridKey == "" {
					sendgridKey = "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"
					log.Println("Warning: Set SENDGRID_KEY variable for production usage.")
					//		return errors.New("Fatal: environmental variable `SENDGRID_KEY` is Crucial.\n\n\t\tHint: export SENDGRID_KEY=123456789")
				}
				if cosgoDestination == "" {
					cosgoDestination = os.Getenv("COSGO_DESTINATION")
					if cosgoDestination == "" {
						return errors.New("Fatal: environmental variable `COSGO_DESTINATION` is Crucial.\n\n\t\tHint: export COSGO_DESTINATION=you@yours.com")
					}
				}
			default:
				// No mailmode, going for mailbox.
				*mailmode = localmail
				mailbox = true
				quickSelfTest()

			}

		} else {
			// Mailbox mode chosen. We aren't really sending any mail so we don't need a real email address to send it to.
			cosgoDestination = os.Getenv("COSGO_DESTINATION")
			if cosgoDestination == "" {
				if !*quiet && !*debug {
					log.Println("Info: COSGO_DESTINATION not set. Using user@example.com")
					log.Println("Hint: export COSGO_DESTINATION=\"your@email.com\"")
				}
			}
		}
	}

	// Main template. Replace with your own, but keep the {{.Tags}} in it.
	_, err = template.New("Index").ParseFiles("./templates/index.html")
	if err != nil {
		log.Println("Fatal: Template Error:", err)
		log.Fatal("Fatal: Template Error\n\n\t\tHint: Copy ./templates and ./static from $GOPATH/src/github.com/aerth/cosgo/ to the location you are running cosgo from.")
	}

	// Make sure Error pages template is present
	_, err = template.New("Error").ParseFiles("./templates/error.html")
	if err != nil {
		log.Println("Fatal: Template Error:", err)
		log.Fatal("Fatal: Template Error\nHint: Copy ./templates and ./static from $GOPATH/src/github.com/aerth/cosgo/ to the location you are running cosgo from.")
	}
	return nil
}

func setDestination() {
	if mailbox == true || *mailmode == localmail {
		return
	}
	cosgoDestination = os.Getenv("COSGO_DESTINATION")
	if cosgoDestination == "" {
		log.Fatal("Fatal: environmental variable `COSGO_DESTINATION` is Crucial.\n\n\t\tHint: export COSGO_DESTINATION=\"your@email.com\"")
		os.Exit(1)
	}
}

//getLink returns the bind:port or http://bind:port string
func getLink(fastcgi bool, showbind string, port string) string {
	if showbind == "0.0.0.0" {
		showbind = "localhost"
	}
	if fastcgi == true {
		link := "fastcgi://" + showbind + ":" + port
		return link
	}
	link := "http://" + showbind + ":" + port
	return link

}

func detectConfig() bool {

	// Detect seconf file. Create if it doesn't exist.
	if !seconf.Detect(*custom) {
		fmt.Println("Config file: " + seconf.Locate(*custom) + " doesn't exist.")
		return false
	}
	return true
}
func loadConfig() bool {

	// Detect seconf file. Create if it doesn't exist.
	if !seconf.Detect(*custom) {
		fmt.Println("Config file: " + seconf.Locate(*custom) + " doesn't exist.")
		conf := seconf.ConfirmChoice("Create it?", true)
		if !conf {
			fmt.Println(conf)
			os.Exit(1)
		}
		fmt.Println("Creating config at " + seconf.Locate(*custom) + ". Don't exit during this.")
		seconf.Create(*custom,
			"\n[cosgo config generator]", // Title
			"\n32 bit CSRF Key, Press ENTER for autogen.( not recommended )\nWill not echo: ",
			"\nCOSGO_KEY: Press ENTER for autogen.( recommended )\nWill not echo: ",
			"\nCOSGO_DESTINATION, Press ENTER for cosgo@localhost",
			"\nPlease select from the following mailbox options. \n\n\t\tmandrill\tsendgrid. \n\nPress ENTER for local mailbox mode (manual reply).",
			"\nMANDRILL_KEY, Press ENTER for local or sendgrid.",
			"\nSENDGRID_KEY, Press ENTER for local or mandrill.",
			"\nWhich port to listen on? Press ENTER for 8080",
			"\nWhich file to use for logs? Press ENTER for cosgo.log",
			"\nWhich interface to listen on? Press ENTER for 0.0.0.0 ( all interfaces )",
			"\nAccept insecure cookie transmission? Press ENTER for yes")
	}

	// Now that a config file exists, unlock it.
	configdecoded, err := seconf.Read(*custom)
	if err != nil {
		fmt.Println("error:")
		fmt.Println(err)
		return false
	}
	configarray := strings.Split(configdecoded, "::::")

	// Cosgo 0.6 uses new config file!
	if len(configarray) < 8 {
		fmt.Println("Broken config file. Create a new one. This happens when updating cosgo version.")
		if seconf.ConfirmChoice("Would you like to delete it? "+seconf.Locate(*custom), true) {
			seconf.Destroy(*custom)
		}
		return loadConfig()
	}
	if err != nil {
		fmt.Println(err)
		return false
	}
	antiCSRFkey = []byte(configarray[0])
	cosgo.PostKey = configarray[1]
	cosgoDestination = configarray[2]
	*mailmode = configarray[3]
	mandrillKey = configarray[4]
	sendgridKey = configarray[5]
	*port = configarray[6]
	*logpath = configarray[7]
	*bind = configarray[8]
	if *logpath == "" {
		*logpath = "cosgo.log"
	}
	if configarray[9] == "no" {
		*secure = true
	}

	if configarray[0] == "" {
		antiCSRFkey = []byte("LI80POC1xcT01jmQBsEyxyrDCrbyyFPjPU8CKnxwmCruxNijgnyb3hXXD3p1RBc0+LIRQUUbTtis6hc6LD4I/A==")
	}

	if cosgo.PostKey == "" {
		cosgo.PostKey = ""
	}

	if cosgoDestination == "" {
		cosgoDestination = "user@example.com"
	}

	if configarray[3] == "" {
		*mailmode = localmail
	}

	if configarray[4] == "" {
		mandrillKey = ""
	}
	if configarray[5] == "" {
		sendgridKey = ""
	}
	if *bind == "" {
		*bind = "0.0.0.0"
	}
	if *bind == "local" {
		*bind = "127.0.0.1"
	}
	if *port == "" {
		*port = "8080"
	}

	// switch *mailmode {
	// case localmail:
	// 	if !*quiet && *debug {
	// 		log.Println("Saving mail (cosgo.mbox) addressed to " + cosgoDestination)
	// 	}
	// case smtpmandrill:
	// 	if !*quiet && *debug {
	// 		log.Println("Sending via Mandrill to " + cosgoDestination)
	// 	}
	// case smtpsendgrid:
	// 	if !*quiet && *debug {
	// 		log.Println("Sending via Sendgrid to " + cosgoDestination)
	// 	}
	// default:
	// 	log.Fatalln("No mailmode.")
	// }

	return true
}

//openLogFile switches the log engine to a file, rather than stdout.
func openLogFile() {
	f, err := os.OpenFile(*logpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("error opening file: %v", err)
		log.Fatal("Hint: touch " + *logpath + ", or chown/chmod it so that the cosgo process can access it.")
		os.Exit(1)
	}
	log.SetOutput(f)
}
