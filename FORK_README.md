# Sup (HCL2 Edition) 🚀

> Ce dépôt est un fork de [pressly/sup](https://github.com/pressly/sup) ajoutant le support natif des fichiers **HCL2**.

`sup` est un outil de déploiement ultra-léger. Ce fork lui permet d'utiliser la syntaxe HCL2 (le langage de Terraform) pour une configuration plus moderne et lisible.

## ✨ Nouveautés

- **Support .hcl** : Chargez vos fichiers de configuration au format HCL2.
- **Architecture Plugin** : Utilise un parseur externe via `go-plugin` pour rester léger et modulaire.

## 📦 Installation

1. **Compiler sup** :
   ```bash
   go build -o sup ./cmd/sup/
   ```

2. **Installer le parseur HCL2** :
   Suivez les instructions sur le dépôt [sup-hcl2-plugin](https://github.com/maelanjais/sup-hcl2-plugin).

## 🚀 Utilisation rapide

Créez un fichier `Supfile.hcl` :

```hcl
version = "0.5"

network "local" {
  hosts = ["localhost"]
}

command "hello" {
  run = "echo 'Hello from HCL2!'"
}
```

Lancez la commande :
```bash
./sup -f Supfile.hcl local hello
```

---

*Développé pour moderniser vos workflows de déploiement.*
