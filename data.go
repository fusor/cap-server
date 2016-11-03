package main

import (
	"fmt"
)

type Answers map[string]map[string]string

// Nulecule
type Nulecule struct {
	Specversion string
	Id          string
	Metadata    map[string]string
	Params      map[string]string
	Graph       []Node
}

type Node struct {
	Name   string
	Source string
	Params []Param
}

type Param struct {
	Name        string
	Description string
	Default     string
	Binds       []string
}

type Bindings struct {
	Src     string `json:"src"`
	SrcKey  string `json:"src_key"`
	Dest    string `json:"dest"`
	DestKey string `json:"dest_key"`
}

// End Nulecule

type NuleculeDetail struct {
	Nulecule Answers    `json:"nulecule"`
	Bindings []Bindings `json:"bindings"`
}

type NuleculeList struct {
	Nulecules []string `json:"nulecules"`
}

// Namespace related types
// This is necessary because we need to strip, and inject back in namespaces
// to be able to be compatible with atomicapp 0.6.4, which introduced the concept.
// Not necessary for 0.6.3
// Example section header atomicapp 0.6.4 expects in the answers file:
//   -> [mariadb-centos7-atomicapp:mariadb-atomicapp]
// 0.6.3 only required [mariadb-atomicapp]
// TODO: Consider a better way to handle this.
type AtomicAppId string

func NewAtomicAppId(registry string, nuleculeId string) AtomicAppId {
	return AtomicAppId(fmt.Sprintf("%s/%s", registry, nuleculeId))
}

// Example namespace: mariadb-centos-atomicapp
// Example name: mariadb-atomicapp
type NamespaceToNameMap map[string]string
type NamespaceManifest map[AtomicAppId]NamespaceToNameMap

func (pManifest *NamespaceManifest) insert(
	registry string,
	nuleculeId string,
	ns string,
	nodeName string,
) NamespaceManifest {
	manifest := *pManifest
	appId := NewAtomicAppId(registry, nuleculeId)
	if mapping, ok := manifest[appId]; ok {
		mapping[ns] = nodeName
	} else {
		newMapping := make(NamespaceToNameMap)
		fmt.Println("inserting ns", ns)
		fmt.Println("inserting nodeName", nodeName)
		fmt.Println("newMapping", newMapping)
		newMapping[ns] = nodeName
		manifest[appId] = newMapping
	}
	return manifest
}
