package main

import (
    "fmt"
    "forensicinvestigator/internal/services"
)

func main() {
    content := `:: suspects ::

@chef_reseau Viktor Sokolov (type) personne

@temoin_cle Claire Fontaine (type) personne

$temoin_cle.1 (a dénoncé:+L) $chef_reseau.1
`
    
    svc := services.NewN4LService()
    result := svc.ParseN4L(content)
    
    fmt.Println("=== Aliases ===")
    for k, v := range result.Aliases {
        fmt.Printf("%s: %v\n", k, v)
    }
    
    fmt.Println("\n=== Subjects ===")
    for _, s := range result.Subjects {
        fmt.Printf("  - %s\n", s)
    }
    
    fmt.Println("\n=== Edges ===")
    for _, e := range result.Graph.Edges {
        fmt.Printf("  %s -[%s]-> %s\n", e.From, e.Label, e.To)
    }
}
