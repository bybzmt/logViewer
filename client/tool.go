package client

import (
	"log"
	"logViewer/find"
	"regexp"
	"regexp/syntax"

	"github.com/timshannon/bolthold"
	"golang.org/x/crypto/ssh"
)

var crossRegexp *regexp.Regexp

func init() {
	reg, err := syntax.Parse("^(http|https)://(127.0.0.1|localhost)(:\\d+)?", syntax.Perl)
	if err != nil {
		log.Panicln(err)
	}

	crossRegexp, err = regexp.Compile(reg.String())
	if err != nil {
		log.Panicln(err)
	}
}

func findDial(store *bolthold.Store, logId uint64) (find.Client, error) {
	var logCfg ViewLog
	err := store.FindOne(&logCfg, bolthold.Where("ID").Eq(logId))
	if err != nil {
		log.Println("FindOne", err)
		return nil, err
	}

	if logCfg.ServerID == 0 {
		return find.NewClient(), nil
	} else if logCfg.ServerID == 2 {
		addr := "./cmd/cli/cli"
		return find.NewClientCLI(addr)
	}

	var serCfg ServerConfig

	err = store.FindOne(&serCfg, bolthold.Where("ID").Eq(logCfg.ServerID))
	if err != nil {
		log.Println("FindOne", err)
		return nil, err
	}

	sshCfg := ssh.ClientConfig{
		User:            serCfg.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if serCfg.UsePwd {
		sshCfg.Auth = []ssh.AuthMethod{
			ssh.Password(serCfg.Passwd),
		}
	} else {
		signer, err := ssh.ParsePrivateKey([]byte(serCfg.PrivateKey))
		if err != nil {
			log.Println("unable to parse private key:", err)
			return nil, err
		}
		sshCfg.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}

	client, err := ssh.Dial("tcp", serCfg.Addr, &sshCfg)
	if err != nil {
		log.Println("ssh Dial", err)
		return nil, err
	}

	addr := "./cmd/cli/cli"

	return find.NewClientSSH(client, addr)
}
