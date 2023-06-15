package iptables

import (
	"crypto/sha256"
	"encoding/base32"

	"github.com/coreos/go-iptables/iptables"
	"github.com/nadoo/ipset"

	"github.com/sirupsen/logrus"
)

const (
	ipsetName = "egressgw"
	mark      = "1234"
)

func GatewayActions(finalPodIPs []string, destinationCIDRs []string) error {
	// Create the tunnel
	logrus.Info("In the future, I will create the tunnel for the gateway")

	// CreateIPSet with the source pods
	// DO I NEED TO CREATE THE IPSET IN THE GATEWAY??
	//ipsetName, err := createIPset(finalPodIPs, destinationCIDRs[0])
	ipt, err := iptables.New()
	if err != nil {
		logrus.Error("There is an error when creating iptables.New()")
		return err
	}

	logrus.Info("About to check if nat/EGRESSGW-MASQ exists")
	exists, err := ipt.Exists("nat", "EGRESSGW-MASQ")
	if err != nil {
		logrus.Error("Error while checking if EGRESSGW-MASQ exists")
		return err
	}

	logrus.Infof("In GatewayActions, this is finalPodIPs: %v and this is destinationCIDRs: %v", finalPodIPs, destinationCIDRs)

	if !exists {
		err = createEgressGwMasqChain(*ipt)
		if err != nil {
			logrus.Error("Error while creating the EgressGwMasq")
		}
	}

	// Create the masquerade rule for the outgoing traffic
	for _, cidr := range destinationCIDRs {
		rule := []string{"-d", cidr, "-j", "MASQUERADE", "-m", "comment", "--comment", "egressGw masq"}
		err = ipt.Insert("nat", "EGRESSGW-MASQ", 1, rule...)
		if err != nil {
			logrus.Error("Error while inserting EGRESSGW-MASQ rule with cidr: " + cidr)
			return err
		}
	}

	return nil
}

func NonGatewayActions(finalPodIPs []string, destinationCIDRs []string) error {

	// Create tunnel
	logrus.Info("In the future, I will create the tunnel for the non-gateway")

	logrus.Infof("In NonGatewayActions, this is finalPodIPs: %v and this is destinationCIDRs: %v", finalPodIPs, destinationCIDRs)

	// Create IPSet with the source pods
	ipsetName, err := createIPset(finalPodIPs, destinationCIDRs[0])
	if err != nil {
		logrus.Error("Error creating the ipset")
		return err
	}

	logrus.Infof("This is ipsetName: %s", ipsetName)

	// Create rule to intercept traffic
	ipt, err := iptables.New()
	if err != nil {
		logrus.Error("Error when creating iptables.New()")
		return err
	}

	exists, err := ipt.Exists("mangle", "EGRESSGW-INTERCEPT")
	if err != nil {
		logrus.Error("Error while checking if EGRESSGW-INTERCEPT exists")
		return err
	}

	if !exists {
		err = createEgressGwInterceptChain(*ipt)
		if err != nil {
			logrus.Error("Error while creating the EgressGwIntercept")
		}
	}

	// Create the interception rule
	for _, cidr := range destinationCIDRs {
		rule := []string{"-d", cidr, "-m", "set", "--match-set", ipsetName, "src", "-j", "MARK", "--set-mark", mark, "-m", "comment", "--comment", "egressGw intercept"}
		err = ipt.Insert("mangle", "EGRESSGW-INTERCEPT", 1, rule...)
		if err != nil {
			logrus.Error("Error while inserting EGRESSGW-INTERCEPT rule with cidr: " + cidr)
			return err
		}
	}

	// Create ip route rule

	return nil
}

// Creates an ipset with sourceips. firstDestCidr is just to create the hash
func createIPset(ips []string, firstDestCidr string) (string, error) {
	hash := sha256.Sum256([]byte(ips[0] + firstDestCidr))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	ipsetName := "egressgw" + encoded[:16]
	logrus.Infof("This is the name of the ipset: %s", ipsetName)

	err := ipset.Create(ipsetName)
	if err != nil {
		logrus.Error("Error while creating the ipset")
		return "", err
	}
	for _, ip := range ips {
		err = ipset.Add(ipsetName, ip)
		if err != nil {
			logrus.Error("Error while adding IPs to the ipset")
			return "", err
		}
	}

	return ipsetName, nil
}

// Creates EGRESSGW-MASQ and places it in the nat/POSTROUTING
func createEgressGwMasqChain(ipt iptables.IPTables) error {

	// Flush EGRESSGW-MASQ chain or recreate it
	err := ipt.ClearChain("nat", "EGRESSGW-MASQ")
	if err != nil {
		logrus.Error("Error while clearChaining the EGRESSGW-MASQ chain")
		return err
	}

	// Insert EGRESSGW-MASQ into iptables flow
	rulePostrouting := []string{"-m", "comment", "--comment", "egressGw masq", "-j", "EGRESSGW-MASQ"}
	err = ipt.Insert("nat", "POSTROUTING", 1, rulePostrouting...)
	if err != nil {
		logrus.Error("Error while inserting the EGRESSGW-MASQ chain")
		return err
	}

	return nil
}

// createEgressGwIntercept creates EGRESSGW-INTERCEPT and places it in the mangle/PREROUTING
func createEgressGwInterceptChain(ipt iptables.IPTables) error {

	// Flush EGRESSGW-INTERCEPT chain or recreate it
	err := ipt.ClearChain("mangle", "EGRESSGW-INTERCEPT")
	if err != nil {
		return err
	}
	// Insert EGRESSGW-INTERCEPT into iptables flow
	rulePrerouting := []string{"-m", "comment", "--comment", "egressGw intercept", "-j", "EGRESSGW-INTERCEPT"}
	err = ipt.Insert("mangle", "PREROUTING", 1, rulePrerouting...)
	if err != nil {
		return err
	}

	return nil

}
