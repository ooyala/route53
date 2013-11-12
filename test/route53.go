package main

import (
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/jessevdk/go-flags"
	"os"
	"route53"
	"time"
)

var r53 *route53.Route53

type Route53Client struct {
	*flags.Parser
}

type ClientOpts struct {
	Debug string `short:"d" long:"debug" description:"verbose debug logging"`
}

type ChangeCommand struct {
	Id string `short:"i" long:"id" description:"change id"`
}

func (c *ChangeCommand) Execute(args []string) error {
	change, err := r53.GetChange(c.Id)
	if err != nil {
		return err
	}
	fmt.Printf("change %s submitted %s is %s", change.Id, change.SubmittedAt, change.Status)
	return nil
}

type ZoneCommand struct {
	Cmd       string
	Id        string `short:"i" long:"id" description:"zone id"`
	Name      string `short:"n" long:"name" description:"zone name"`
	Reference string `short:"r" long:"reference" description:"caller reference"`
	Comment   string `short:"c" long:"comment" description:"comment string"`
}

func (z *ZoneCommand) Execute(args []string) error {
	switch z.Cmd {
	case "list-zones":
		r53.ListHostedZones()
	case "get-zone":
		if z.Id == "" {
			fmt.Fprintln(os.Stderr, "error: no id specified")
			os.Exit(255)
		}
		r53.GetHostedZone(z.Id)
	case "add-zone":
		r53.CreateHostedZone(z.Name, z.Reference, z.Comment)
	case "delete-zone":
		r53.DeleteHostedZone(z.Id)
	default:
		fmt.Fprintln(os.Stderr, "error: unknown zone command")
		os.Exit(255)
	}
	return nil
}

type RRSetCommand struct {
	Cmd           string
	ZoneId        string   `short:"z" long:"zone" description:"zone to modify"`
	Comment       string   `short:"c" long:"comment" description:"comment change"`
	Name          string   `short:"n" long:"name" description:"dns domain name"`
	Type          string   `short:"t" long:"type" description:"dns record type"`
	TTL           uint     `short:"l" long:"ttl" description:"record TTL"`
	Values        []string `short:"v" long:"value" description:"resource value"`
	HealthChecks  []string `short:"h" long:"check" description:"health check id"`
	SetIdentifier string   `short:"i" long:"id" description:"record identifier"`

	// Weight Syntax
	Weight uint8 `long:"weight" description:"record weight [0-255]"`

	// Failover Syntax
	FailOver string `long:"failover" description:"primary or secondary"`

	// Latency Syntax
	Region string `long:"region" description:"ec2 region name"`
}

func (r *RRSetCommand) Execute(args []string) error {
	rrset := route53.RRSet{
		Name:          r.Name,
		Type:          r.Type,
		TTL:           r.TTL,
		Values:        r.Values,
		SetIdentifier: r.SetIdentifier,
		Weight:        r.Weight,
		FailOver:      r.FailOver,
		Region:        r.Region,
	}

	switch r.Cmd {
	case "list-rrsets":
		r53.ListRRSets(r.ZoneId)
	case "add-rrset":
		change := route53.RRSetChange{
			Action: "CREATE",
			RRSet:  rrset,
		}
		r53.ChangeRRSet(r.ZoneId, []route53.RRSetChange{change}, r.Comment)
	case "del-rrset":
		change := route53.RRSetChange{
			Action: "DELETE",
			RRSet:  rrset,
		}
		r53.ChangeRRSet(r.ZoneId, []route53.RRSetChange{change}, r.Comment)
	default:
		fmt.Fprintln(os.Stderr, "error: unknown rrset command")
		os.Exit(255)
	}
	return nil
}

type HealthCheckCommand struct {
	Cmd          string
	Id           string `short:"i" long:"id" description:"health check ID"`
	IpAddr       string `short:"a" long:"address" description:"IP address"`
	Port         uint16 `short:"p" long:"port" description:"TCP port"`
	Type         string `short:"t" long:"type" description:"TCP or HTTP"`
	ResourcePath string `long:"path" description:"path for HTTP check"`
	FQDN         string `short:"f" long:"fqdn" description:"FQDN of endpoint"`
	Reference    string `short:"r" long:"reference" description:"caller reference"`
}

func (c *HealthCheckCommand) Execute(args []string) error {
	config := route53.HealthCheckConfig{
		IPAddress:                c.IpAddr,
		Port:                     c.Port,
		Type:                     c.Type,
		ResourcePath:             c.ResourcePath,
		FullyQualifiedDomainName: c.FQDN,
	}

	switch c.Cmd {
	case "list-checks":
		r53.ListHealthChecks()
	case "get-check":
		r53.GetHealthCheck(c.Id)
	case "add-check":
		r53.CreateHealthCheck(config, c.Reference)
	case "delete-check":
		r53.DeleteHealthCheck(c.Id)
	default:
		fmt.Fprintln(os.Stderr, "error: unknown check command")
		os.Exit(255)
	}

	return nil
}

func NewClient() *Route53Client {
	c := &Route53Client{flags.NewParser(nil, flags.Default)}

	c.AddCommand("get-change", "get information for change Id", "", &ChangeCommand{})

	c.AddCommand("list-zones", "list route53 zones", "", &ZoneCommand{Cmd: "list-zones"})
	c.AddCommand("get-zone", "inspect route53 zone", "", &ZoneCommand{Cmd: "get-zone"})
	c.AddCommand("add-zone", "add zone to route53", "", &ZoneCommand{Cmd: "add-zone"})
	c.AddCommand("delete-zone", "delete zone from route53", "", &ZoneCommand{Cmd: "delete-zone"})

	c.AddCommand("list-rrsets", "list resource record sets", "", &RRSetCommand{Cmd: "list-rrsets"})
	c.AddCommand("add-rrset", "add resource record to zone", "", &RRSetCommand{Cmd: "add-rrset"})
	c.AddCommand("delete-rrset", "delete resource record from zone", "", &RRSetCommand{Cmd: "delete-rrset"})

	c.AddCommand("list-checks", "list health checks", "", &HealthCheckCommand{Cmd: "list-checks"})
	c.AddCommand("get-check", "inspect health check", "", &HealthCheckCommand{Cmd: "get-check"})
	c.AddCommand("add-check", "add health check", "", &HealthCheckCommand{Cmd: "add-check"})
	c.AddCommand("delete-check", "delete health check", "", &HealthCheckCommand{Cmd: "delete-check"})

	return c
}

func main() {
	route53.DebugOn()

	auth, err := aws.GetAuth("", "", "", time.Time{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: no aws credentials available")
		os.Exit(255)
	}

	r53 = route53.New(auth)

	NewClient().Parse()
}
