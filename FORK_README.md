# Sup (HCL2 Distributed Edition)

Ce dépôt constitue un fork de l'outil [pressly/sup](https://github.com/pressly/sup). Il introduit une évolution majeure dans la gestion de la configuration en intégrant le support natif du format **HCL2**.

Sup est un outil d'orchestration minimaliste permettant d'exécuter des commandes de manière séquentielle ou parallèle sur des flottes de serveurs via SSH. Cette version enrichie permet d'utiliser le langage de configuration standard de l'industrie (utilisé par Terraform) pour définir vos environnements de déploiement.

## Avantages de l'intégration HCL2

L'adoption du format HCL2 apporte plusieurs bénéfices techniques par rapport au YAML traditionnel :

- **Rigueur syntaxique** : Validation plus stricte des blocs de configuration (networks, commands, targets).
- **Lisibilité accrue** : Structure de données plus naturelle pour la définition d'infrastructures.
- **Extensibilité** : Utilisation de l'architecture `go-plugin` de HashiCorp, permettant de mettre à jour le parseur sans modifier le coeur de Sup.

## Architecture technique

Le support HCL2 ne surcharge pas le binaire principal de Sup. L'intégration repose sur un système de plugin RPC (Remote Procedure Call) :

1. **Détection dynamique** : Sup identifie l'extension du fichier de configuration fourni.
2. **Processus Isolé** : Si le format est HCL2, Sup lance le binaire `sup-hcl2-parser` en tant que processus fils.
3. **Traduction Transparente** : Le plugin traduit la configuration HCL2 en un flux de données compatible avec les structures internes de Sup, garantissant une stabilité totale du moteur d'exécution original.

## Installation

### Compilation du binaire principal
```bash
# Cloner le dépôt
git clone https://github.com/maelanjais/sup.git
cd sup

# Compiler l'exécutable
go build -o sup ./cmd/sup/
```

### Installation du parseur requis
Le fonctionnement de cette version nécessite la présence du binaire `sup-hcl2-parser` dans votre système. Référez-vous au dépôt du plugin pour sa compilation : [sup-hcl2-plugin](https://github.com/maelanjais/sup-hcl2-plugin).

## Exemple de configuration (Supfile.hcl)

```hcl
version = "0.5"

# Définition des variables globales
env {
  PROJECT = "monitoring-stack"
}

# Configuration du réseau
network "cluster" {
  hosts = ["admin@node1.internal", "admin@node2.internal"]
}

# Définition des tâches
command "update" {
  desc = "Mise à jour des services"
  run  = "docker compose pull && docker compose up -d"
}
```

Exécution :
```bash
./sup -f Supfile.hcl cluster update
```

## Roadmap et Maintenance

Ce fork est maintenu activement pour garantir la compatibilité avec les dernières versions de Go et des bibliothèques HashiCorp. Les prochaines versions visent à supporter les fonctions natives HCL (variables, boucles simples) pour une orchestration encore plus dynamique.
