package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

// Enum SNMPType, SNMP_OID, etc. (como no exemplo anterior)
type SNMPType int

const (
	Integer SNMPType = iota
	OctetString
	Null
	OID
	IPAddress
	Counter32
	Gauge32
	TimeTicks
	Opaque
	Counter64
)

func (s SNMPType) String() string {
	return [...]string{"Integer", "OctetString", "Null", "OID", "IPAddress", "Counter32", "Gauge32", "TimeTicks", "Opaque", "Counter64"}[s]
}

type SNMP_OID struct {
	OID   string
	Type  SNMPType
	Value interface{}
}

// Estrutura Node para a árvore de OIDs
type Node struct {
	Name     string
	Type     *SNMPType
	Value    *interface{}
	Children map[string]*Node
}

// Cria um novo nó da árvore
func NewNode(name string) *Node {
	return &Node{
		Name:     name,
		Children: make(map[string]*Node),
	}
}

// Insere uma OID na árvore e imprime conforme insere
func (n *Node) Insert(oidParts []string, snmpType SNMPType, value interface{}, fullOID string) {
	if len(oidParts) == 0 {
		n.Type = &snmpType
		n.Value = &value
		// Imprime a OID completa, o tipo e o valor
		fmt.Printf("OID: %s - Type: %s, Value: %v\n", fullOID, snmpType.String(), value)
		return
	}

	part := oidParts[0]
	child, exists := n.Children[part]
	if !exists {
		child = NewNode(part)
		n.Children[part] = child
	}

	child.Insert(oidParts[1:], snmpType, value, fullOID)
}

// Função principal
func main() {
	// Configuração de conexão SNMP
	snmp := &gosnmp.GoSNMP{
		Target:    "10.100.73.2",  // Endereço IP do dispositivo
		Port:      161,            // Porta padrão SNMP
		Community: "intelbras123", // Comunidade SNMP
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(2) * time.Second,
		Retries:   1,
	}

	err := snmp.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer snmp.Conn.Close()

	// Cria a raiz da árvore
	root := NewNode("")

	// OID inicial para iniciar o "walk"
	oid := ".1.3.6.1.2.1" // OID base para varredura

	// Realiza a operação SNMP GET-NEXT sequencialmente
	for {
		result, err := snmp.GetNext([]string{oid})
		if err != nil {
			log.Fatalf("Erro no GET-NEXT: %v", err)
		}

		if len(result.Variables) == 0 {
			fmt.Println("Nenhuma OID encontrada")
			break
		}

		for _, variable := range result.Variables {
			// Define o tipo de SNMP com base no tipo recebido
			var snmpType SNMPType
			switch variable.Type {
			case gosnmp.Integer:
				snmpType = Integer
			case gosnmp.OctetString:
				snmpType = OctetString
			case gosnmp.IPAddress:
				snmpType = IPAddress
			case gosnmp.Counter32:
				snmpType = Counter32
			case gosnmp.Gauge32:
				snmpType = Gauge32
			case gosnmp.TimeTicks:
				snmpType = TimeTicks
			case gosnmp.Counter64:
				snmpType = Counter64
			default:
				snmpType = Null
			}

			// Divide a OID em partes e insere na árvore, exibindo a OID completa
			oidParts := strings.Split(variable.Name, ".")
			root.Insert(oidParts[1:], snmpType, variable.Value, variable.Name) // Remove o primeiro elemento vazio
			oid = variable.Name                                                // Atualiza a OID para o próximo GET-NEXT
		}
	}
}
