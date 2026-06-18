package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type TargetPreset struct {
	OrgID   string `json:"org_id"`
	Host    string `json:"host"`
	Referer string `json:"referer"`
}

var builtinTargets = map[string]TargetPreset{
	"direct":        {OrgID: "usllpic0", Host: "h.online-metrix.net", Referer: "https://www.example.com/"},
	"walmart":       {OrgID: "hgy2n0ks", Host: "drfdisvc.walmart.com", Referer: "https://identity.walmart.com/"},
	"ebay":          {OrgID: "usllpic0", Host: "src.ebay-us.com", Referer: "https://signin.ebay.com/"},
	"kleinanzeigen": {OrgID: "usllpic0", Host: "dfme.kleinanzeigen.de", Referer: "https://www.kleinanzeigen.de/"},
	"coinbase":      {OrgID: "k8vif92e", Host: "h.online-metrix.net", Referer: "https://www.coinbase.com/"},
}

var customTargets = map[string]TargetPreset{}

func loadCustomTargets(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	parsed := map[string]TargetPreset{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	for k, v := range parsed {
		customTargets[strings.ToLower(k)] = v
	}
	return nil
}

func applyTargetPreset(name string, org, host, referer *string) {
	n := strings.ToLower(name)
	for _, alias := range []string{"ka", "ebay-de"} {
		if n == alias {
			n = "kleinanzeigen"
		}
	}
	if p, ok := customTargets[n]; ok {
		applyPreset(p, org, host, referer)
		return
	}
	if p, ok := builtinTargets[n]; ok {
		applyPreset(p, org, host, referer)
	}
}

func applyPreset(p TargetPreset, org, host, referer *string) {
	if p.OrgID != "" && (*org == "usllpic0" || *org == "") {
		*org = p.OrgID
	}
	if p.Host != "" && (*host == "h.online-metrix.net" || *host == "") {
		*host = p.Host
	}
	if p.Referer != "" && (*referer == "https://www.example.com/" || *referer == "") {
		*referer = p.Referer
	}
}
