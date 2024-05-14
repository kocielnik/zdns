package dmarc

import (
	"errors"
	"github.com/spf13/pflag"
	"github.com/zmap/dns"
	"github.com/zmap/zdns/cmd"
	"github.com/zmap/zdns/pkg/zdns"
	"regexp"
)

const dmarcPrefixRegexp = "^[vV][\x09\x20]*=[\x09\x20]*DMARC1[\x09\x20]*;[\x09\x20]*"

// result to be returned by scan of host
type Result struct {
	Dmarc string `json:"dmarc,omitempty" groups:"short,normal,long,trace"`
}

func init() {
	dmarc := new(DmarcLookupModule)
	cmd.RegisterLookupModule("DMARC", dmarc)
}

type DmarcLookupModule struct {
	cmd.BasicLookupModule
	re *regexp.Regexp
}

func (dmarcMod *DmarcLookupModule) CLIInit(gc *cmd.CLIConf, rc *zdns.ResolverConfig, flags *pflag.FlagSet) error {
	dmarcMod.re = regexp.MustCompile(dmarcPrefixRegexp)
	dmarcMod.BasicLookupModule.DNSType = dns.TypeTXT
	dmarcMod.BasicLookupModule.DNSClass = dns.ClassINET
	return dmarcMod.BasicLookupModule.CLIInit(gc, rc, flags)
}

func (dmarcMod *DmarcLookupModule) Lookup(r *zdns.Resolver, lookupName, nameServer string) (interface{}, zdns.Trace, zdns.Status, error) {
	innerRes, trace, status, err := dmarcMod.BasicLookupModule.Lookup(r, lookupName, nameServer)
	castedInnerRes, ok := innerRes.(*zdns.SingleQueryResult)
	if !ok {
		return nil, trace, status, errors.New("lookup didn't return a single query result type")
	}
	resString, resStatus, err := zdns.CheckTxtRecords(castedInnerRes, status, dmarcMod.re, err)
	res := Result{Dmarc: resString}
	return res, trace, resStatus, err
}

// Help
func (dmarcMod *DmarcLookupModule) Help() string {
	return ""
}
