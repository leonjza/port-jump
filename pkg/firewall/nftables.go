//go:build linux

package firewall

import (
	"fmt"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

const (
	tableName = "port-jump"
	chainName = "prerouting"
)

// AddOrUpdateRedirect updates the firewall using NFTables to redirect traffic from, to.
func AddOrUpdateRedirect(from int, to int) error {
	conn := &nftables.Conn{}

	// Get or create the NAT table
	table, err := getOrCreateTable(conn, tableName, nftables.TableFamilyIPv4)
	if err != nil {
		return fmt.Errorf("Failed to get or create table: %v", err)
	}

	// Get or create the chain
	chain, err := getOrCreateChain(conn, table, chainName)
	if err != nil {
		return fmt.Errorf("Failed to get or create chain: %v", err)
	}

	// Find the existing rule and update it if needed
	err = findAndUpdateRule(conn, table, chain, from, to)
	if err != nil {
		return fmt.Errorf("Failed to update rule: %vn", err)
	}

	return nil
}

// DeleteRules deletes any created rules by deleting the custom table created
func DeleteRules() error {
	conn := &nftables.Conn{}

	tables, err := conn.ListTables()
	if err != nil {
		return fmt.Errorf("failed to list tables: %v", err)
	}

	var tableToDelete *nftables.Table
	for _, table := range tables {
		if table.Name == tableName {
			tableToDelete = table
			break
		}
	}

	if tableToDelete == nil {
		return fmt.Errorf("table %s not found", tableName)
	}

	conn.DelTable(tableToDelete)

	// Apply the changes
	if err := conn.Flush(); err != nil {
		return fmt.Errorf("failed to delete table %s: %v", tableName, err)
	}

	return nil
}

// findAndUpdateRule finds an existing NAT rule by destination port and updates it with a new source port if needed
func findAndUpdateRule(conn *nftables.Conn, table *nftables.Table, chain *nftables.Chain, newSrcPort, targetPort int) error {
	rules, err := conn.GetRules(table, chain)
	if err != nil {
		return fmt.Errorf("failed to get rules: %v", err)
	}

	for _, rule := range rules {
		if ruleMatches(rule, targetPort) {
			if err := conn.DelRule(rule); err != nil {
				return fmt.Errorf("failed to delete existing rule: %v", err)
			}
		}
	}

	return addRedirectRule(conn, table, chain, newSrcPort, targetPort)
}

// ruleMatches checks if a rule matches the given source port
func ruleMatches(rule *nftables.Rule, dstPort int) bool {
	targetMatched := false
	redirMatched := false

	for _, e := range rule.Exprs {
		switch expr := e.(type) {
		case *expr.Redir:
			redirMatched = true
		case *expr.Immediate:
			if len(expr.Data) == 2 {
				ruleTargetPort := int(expr.Data[0])<<8 | int(expr.Data[1])
				if ruleTargetPort == dstPort {
					targetMatched = true
				}
			}
		}

	}

	return targetMatched && redirMatched
}

// addRedirectRule adds a new NAT redirect rule to the chain
func addRedirectRule(conn *nftables.Conn, table *nftables.Table, chain *nftables.Chain, srcPort, targetPort int) error {
	conn.AddRule(&nftables.Rule{
		Table: table,
		Chain: chain,
		Exprs: []expr.Any{
			// Match TCP packets
			&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.IPPROTO_TCP},
			},
			// Match packets destined for srcPort
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2, // 2 bytes offset to get the destination port in TCP/UDP headers
				Len:          2, // Port is 2 bytes long
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{byte(srcPort >> 8), byte(srcPort & 0xff)},
			},
			// Redirect to targetPort (e.g., SSH on port 22)
			&expr.Immediate{
				Register: 1,
				Data:     []byte{0, byte(targetPort)},
			},
			&expr.Redir{
				RegisterProtoMin: 1,
			},
		},
	})

	// Apply the changes
	if err := conn.Flush(); err != nil {
		return fmt.Errorf("failed to add redirect rule: %v", err)
	}

	return nil
}

// getOrCreateTable checks if a table exists, and creates it if it doesn't
func getOrCreateTable(conn *nftables.Conn, tableName string, family nftables.TableFamily) (*nftables.Table, error) {
	tables, err := conn.ListTables()
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %v", err)
	}

	for _, tbl := range tables {
		if tbl.Name == tableName && tbl.Family == family {
			return tbl, nil
		}
	}

	// Table doesn't exist, so create it
	table := conn.AddTable(&nftables.Table{
		Family: family,
		Name:   tableName,
	})

	if err := conn.Flush(); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return table, nil
}

// getOrCreateChain checks if a chain exists in the specified table, and creates it if it doesn't
func getOrCreateChain(conn *nftables.Conn, table *nftables.Table, chainName string) (*nftables.Chain, error) {
	chains, err := conn.ListChains()
	if err != nil {
		return nil, fmt.Errorf("failed to list chains: %v", err)
	}

	for _, chn := range chains {
		if chn.Name == chainName && chn.Table.Name == table.Name {
			return chn, nil
		}
	}

	// Chain doesn't exist, so create it
	chain := conn.AddChain(&nftables.Chain{
		Name:     chainName,
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityNATDest,
	})

	if err := conn.Flush(); err != nil {
		return nil, fmt.Errorf("failed to create chain: %v", err)
	}

	return chain, nil
}
