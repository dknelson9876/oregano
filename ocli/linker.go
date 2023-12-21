package ocli

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/plaid/plaid-go/plaid"
	"github.com/skratchdot/open-golang/open"
)

type Linker struct {
	Results       chan string
	RelinkResults chan bool
	Errors        chan error
	Client        *plaid.APIClient
	Data          *Data
	countries     []string
	lang          string
}

type TokenPair struct {
	ItemID      string
	AccessToken string
}

func NewLinker(data *Data, client *plaid.APIClient, countries []string, lang string) *Linker {
	return &Linker{
		Results:       make(chan string),
		RelinkResults: make(chan bool),
		Errors:        make(chan error),
		Client:        client,
		Data:          data,
		countries:     countries,
		lang:          lang,
	}
}

func (l *Linker) Link(port string) (*TokenPair, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: hostname,
	}
	request := plaid.NewLinkTokenCreateRequest(
		"oregano-cli",
		l.lang,
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
		user,
	)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})
	request.SetLinkCustomizationName("default")
	// request.SetWebhook("https://webhook-uri.com")
	// request.SetRedirectUri("https://your-domain.com/oauth-page.html")
	resp, _, err := l.Client.PlaidApi.LinkTokenCreate(nil).LinkTokenCreateRequest(*request).Execute()

	return l.link(port, resp.GetLinkToken())
}

func (l *Linker) link(port string, linkToken string) (*TokenPair, error) {
	log.Println(fmt.Sprintf("Starting Plaid Link on port %s...", port))

	go func() {
		http.HandleFunc("/link", handleLink(l, linkToken))
		err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
		if err != nil {
			l.Errors <- err
		}
	}()

	url := fmt.Sprintf("http://localhost:%s/link", port)
	log.Println(fmt.Sprintf("Your browser should open automatically. If it doesn't, please visit %s to continue linking!", url))
	open.Run(url)

	select {
	case err := <-l.Errors:
		return nil, err
	case publicToken := <-l.Results:
		res, err := l.exchange(publicToken)
		if err != nil {
			return nil, err
		}

		pair := &TokenPair{
			ItemID:      res.ItemId,
			AccessToken: res.AccessToken,
		}

		return pair, nil
	}
}

func (l *Linker) exchange(publicToken string) (plaid.ItemPublicTokenExchangeResponse, error) {
	// return l.Client.PlaidApi.ItemPublicTokenExchange(nil)
	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	exchangePublicTokenResp, _, err := l.Client.PlaidApi.ItemPublicTokenExchange(nil).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	return exchangePublicTokenResp, err
}

func handleLink(linker *Linker, linkToken string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			t := template.New("link")
			t, _ = t.Parse(linkTemplate)

			d := LinkTmplData{
				LinkToken: linkToken,
			}
			t.Execute(w, d)
		case http.MethodPost:
			r.ParseForm()
			token := r.Form.Get("public_token")
			if token != "" {
				linker.Results <- token
			} else {
				linker.Errors <- errors.New("empty public_token")
			}

			fmt.Fprintf(w, "ok")
		default:
			linker.Errors <- errors.New("invalid HTTP method")
		}
	}
}

type LinkTmplData struct {
	LinkToken string
}

var linkTemplate string = `<html>
  <head>
    <style>
    .alert-success {
	font-size: 1.2em;
	font-family: Arial, Helvetica, sans-serif;
	background-color: #008000;
	color: #fff;
	display: flex;
	justify-content: center;
	align-items: center;
	border-radius: 15px;
	width: 100%;
	height: 100%;
    }
    .hidden {
	visibility: hidden;
    }
    </style>
  </head>
  <body>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.2.3/jquery.min.js"></script>
    <script src="https://cdn.plaid.com/link/v2/stable/link-initialize.js"></script>
    <script type="text/javascript">
     (function($) {
       var handler = Plaid.create({
	 token: '{{ .LinkToken }}',
	 onSuccess: function(public_token, metadata) {
	   // Send the public_token to your app server.
	   // The metadata object contains info about the institution the
	   // user selected and the account ID or IDs, if the
	   // Select Account view is enabled.
	   $.post('/link', {
	     public_token: public_token,
	   });
	   document.getElementById("alert").classList.remove("hidden");
	 },
	 onExit: function(err, metadata) {
	   // The user exited the Link flow.
	   if (err != null) {
	     // The user encountered a Plaid API error prior to exiting.
	   }
	   // metadata contains information about the institution
	   // that the user selected and the most recent API request IDs.
	   // Storing this information can be helpful for support.

	   document.getElementById("alert").classList.remove("hidden");
	 }
       });

       handler.open();

     })(jQuery);
    </script>

    <div id="alert" class="alert-success hidden">
      <div>
	<h2>All done here!</h2>
	<p>You can close this window and go back to oregano-cli.</p>
      </div>
    </div>
  </body>
</html> `
