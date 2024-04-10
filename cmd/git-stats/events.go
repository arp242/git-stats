package main

import (
	"strings"

	gitstats "zgo.at/git-stats"
	"zgo.at/zstd/ztime"
)

// A list of known events that we can't get from git
var knownEvents = []struct {
	repos  []string
	Events gitstats.Events
}{
	{[]string{"https://github.com/hashicorp/terraform", "https://github.com/opentofu/opentofu"}, gitstats.Events{
		{Date: ztime.FromString("2023-08-10"), Kind: 'l', Name: "MPL → BSL"},
		{Date: ztime.FromString("2023-08-16"), Kind: 'f', Name: "Terraform → OpenTofu"},
	}},
	{[]string{"https://github.com/mysql/mysql-server", "https://github.com/MariaDB/server"}, gitstats.Events{
		{Date: ztime.FromString("2008-01-16"), Kind: 'o', Name: "MySQL AB → Sun"},
		{Date: ztime.FromString("2010-01-27"), Kind: 'o', Name: "Sun → Oracle"},
		{Date: ztime.FromString("2010-01-27"), Kind: 'f', Name: "MySQL → MariaDB"},
	}},
	{[]string{"https://github.com/apache/openoffice", "https://git.libreoffice.org/core"}, gitstats.Events{
		{Date: ztime.FromString("2010-01-27"), Kind: 'o', Name: "Sun → Oracle"},
		{Date: ztime.FromString("2011-01-25"), Kind: 'f', Name: "OpenOffice.org → Libreoffice"},
		{Date: ztime.FromString("2012-05-08"), Kind: 'o', Name: "Oracle → Apache"},
	}},
	{[]string{"https://github.com/elastic/elasticsearch", "https://github.com/opensearch-project/opensearch"}, gitstats.Events{
		{Date: ztime.FromString("2021-02-03"), Kind: 'l', Name: "Apache 2.0 → SPPL"},
		{Date: ztime.FromString("2021-04-12"), Kind: 'f', Name: "Elasticsearch → OpenSearch"},
	}},
	{[]string{"https://github.com/redis/redis"}, gitstats.Events{
		{Date: ztime.FromString("2024-03-20"), Kind: 'l', Name: "BSD-3 → RSAL 2.0 / SSPL"},
	}},
	{[]string{"https://github.com/mongodb/mongo"}, gitstats.Events{
		{Date: ztime.FromString("2018-10-16"), Kind: 'l', Name: "AGPL 3 → SSPL"},
	}},
	{[]string{"https://github.com/netbsd/src", "https://github.com/openbsd/src"}, gitstats.Events{
		{Date: ztime.FromString("1995-10-18"), Kind: 'f', Name: "NetBSD → OpenBSD"},
	}},
	{[]string{"https://github.com/freebsd/freebsd-src", "https://gitstats.arp242.net/dragonflybsd"}, gitstats.Events{
		{Date: ztime.FromString("2003-07-16"), Kind: 'f', Name: "FreeBSD → DragonFlyBSD"},
	}},
	{[]string{"https://github.com/getsentry/sentry"}, gitstats.Events{
		{Date: ztime.FromString("2023-11-17"), Kind: 'l', Name: "BSL 1.1 → FSL 1.0"},
		{Date: ztime.FromString("2022-02-27"), Kind: 'l', Name: "FSL 1.0 → FSL 1.1"},
		{Date: ztime.FromString("2019-11-06"), Kind: 'l', Name: "BSD-3 → BSL 1.1"},
	}},
	{[]string{"https://github.com/vim/vim", "https://github.com/neovim/neovim"}, gitstats.Events{
		{Date: ztime.FromString("2015-11-01"), Kind: 'f', Name: "Vim → NeoVim"},
		{Date: ztime.FromString("2023-08-03"), Kind: 'o', Name: "Bram's death"},
	}},
	{[]string{"https://github.com/nginx/nginx"}, gitstats.Events{
		{Date: ztime.FromString("2019-03-11"), Kind: 'o', Name: "Nginx, Inc → F5"},
	}},

	{[]string{"https://github.com/go-gitea/gitea", "https://codeberg.org/forgejo/forgejo", "https://github.com/gogs/gogs"}, gitstats.Events{
		{Date: ztime.FromString("2016-12-06"), Kind: 'f', Name: "Gogs → Gitea"},
		{Date: ztime.FromString("2022-11-26"), Kind: 'f', Name: "Gitea → Forgejo"},
	}},
}

func findKnown(path string) gitstats.Events {
	path = strings.ToLower(path)
	for _, e := range knownEvents {
		for _, r := range e.repos {
			if r == path {
				return e.Events
			}
		}
	}
	return nil
}
