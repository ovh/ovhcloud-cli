# Spécification : Commande `ovhcloud browser`

## 📋 Vue d'ensemble

La commande `ovhcloud browser` lance une interface utilisateur en terminal (TUI) qui simule le Manager OVHcloud. Elle permet de naviguer visuellement à travers les ressources cloud directement depuis le terminal.

---

## 📁 Structure des fichiers

| Fichier | Rôle |
|---------|------|
| [internal/cmd/browser.go](../internal/cmd/browser.go) | Définition de la commande Cobra |
| [internal/services/browser/manager.go](../internal/services/browser/manager.go) | Logique principale de la TUI (~790 lignes) |
| [internal/services/browser/api.go](../internal/services/browser/api.go) | Appels API pour récupérer les données (~610 lignes) |

---

## 🎯 Fonctionnalités implémentées

### Flux de navigation

```
┌─────────────────────────────────────────────────────────────────┐
│  1. DÉMARRAGE → Liste des projets (ProjectSelectView)           │
│     ↑↓: Naviguer • Enter: Sélectionner • d: Set Default         │
├─────────────────────────────────────────────────────────────────┤
│  2. PROJET SÉLECTIONNÉ → Produits du projet                     │
│     Barre de navigation: Instances | Kubernetes | Databases...  │
│     ←→: Changer de produit • Enter: Détails • p: Changer projet │
├─────────────────────────────────────────────────────────────────┤
│  3. VUE DÉTAIL → Informations détaillées d'une ressource        │
│     Esc: Retour à la liste • p: Changer de projet               │
└─────────────────────────────────────────────────────────────────┘
```

### Modes de vue (ViewMode)

| Mode | Description |
|------|-------------|
| `ProjectSelectView` | Sélection du projet (écran d'accueil) |
| `TableView` | Liste des ressources d'un produit |
| `DetailView` | Détails d'une ressource sélectionnée |
| `LoadingView` | Chargement en cours |
| `ErrorView` | Affichage d'une erreur |

### Produits navigables (ProductType)

| Produit | Endpoint API | Icône |
|---------|--------------|-------|
| Instances | `/v1/cloud/project/{id}/instance` | 💻 |
| Kubernetes | `/v1/cloud/project/{id}/kube` | ☸️ |
| Databases | `/v1/cloud/project/{id}/database/service` | 🗄️ |
| Storage | `/v1/cloud/project/{id}/storage/s3` | 💾 |
| Networks | `/v1/cloud/project/{id}/network/private` | 🌐 |
| Projects | `/v1/cloud/project` | 📦 |

### Raccourcis clavier

| Mode | Touche | Action |
|------|--------|--------|
| **Sélection projet** | `↑↓` | Naviguer dans la liste |
| | `Enter` | Sélectionner le projet et afficher ses produits |
| | `d` | Définir comme projet par défaut (sauvegarde dans config) |
| | `q` | Quitter |
| **Liste produits** | `←→` | Changer de produit (Instances ↔ K8s ↔ DB...) |
| | `↑↓` | Naviguer dans la liste |
| | `Enter` | Voir les détails de la ressource |
| | `p` | Retourner à la sélection de projet |
| | `q` | Quitter |
| **Vue détail** | `Esc` | Retour à la liste |
| | `p` | Retourner à la sélection de projet |
| | `q` | Quitter |

### Fonctionnalités spéciales

#### ✅ Set default project (touche `d`)

Depuis la vue de sélection de projet, appuyer sur `d` sauvegarde le projet sélectionné comme `default_cloud_project` dans la configuration CLI (`~/.ovh.conf`).

**Implémentation** :
- Appelle `config.SetConfigValue()` avec `flags.CliConfigPath`
- Affiche une notification temporaire (3 secondes) : `✅ Default project set to: <nom>`
- Met à jour le projet courant dans le browser

#### ✅ Gestion des requêtes obsolètes

Quand l'utilisateur change d'onglet avant la fin du chargement, les réponses de l'ancienne requête sont ignorées.

**Implémentation** :
- Chaque message (`dataLoadedMsg`, `projectsLoadedMsg`) contient un champ `forProduct`
- Les handlers comparent `msg.forProduct` avec `m.currentProduct`
- Si différent, la réponse est ignorée

#### ✅ Cache des projets

La liste des projets est mise en cache dans `Model.projectsList` pour permettre un retour rapide à la sélection de projet (touche `p`) sans refaire d'appel API.

---

## 🏗️ Structure du code

### Model (état de l'application)

```go
type Model struct {
    width              int
    height             int
    mode               ViewMode
    currentProduct     ProductType
    navIdx             int                      // Index dans la barre de navigation
    table              table.Model              // Tableau bubbletea
    detailData         map[string]interface{}   // Données de la ressource en détail
    currentData        []map[string]interface{} // Données de la liste courante
    errorMsg           string
    cloudProject       string                   // ID du projet sélectionné
    cloudProjectName   string                   // Nom du projet sélectionné
    currentItemName    string                   // Nom de l'item en vue détail
    notification       string                   // Message de notification temporaire
    notificationExpiry time.Time                // Expiration de la notification
    projectsList       []map[string]interface{} // Cache des projets
}
```

### Messages asynchrones

```go
type projectsLoadedMsg struct {
    projects   []map[string]interface{}
    err        error
    forProduct ProductType  // Pour ignorer les réponses obsolètes
}

type dataLoadedMsg struct {
    data       []map[string]interface{}
    err        error
    forProduct ProductType  // Pour ignorer les réponses obsolètes
}

type setDefaultProjectMsg struct {
    projectID   string
    projectName string
    err         error
}

type clearNotificationMsg struct{}  // Pour effacer la notification après timeout
```

---

## 🚀 Évolutions futures possibles

### Options CLI à ajouter

| Option | Description | Priorité |
|--------|-------------|----------|
| `--cloud-project` | Pré-sélectionner un projet au démarrage | Moyenne |
| `--start-view` | Démarrer sur un produit spécifique | Basse |
| `--no-color` | Désactiver les couleurs | Basse |
| `--refresh-interval` | Rafraîchissement automatique | Basse |

### Fonctionnalités TUI à ajouter

| Fonctionnalité | Description | Priorité |
|----------------|-------------|----------|
| Recherche/filtre | Filtrer les ressources par nom | Moyenne |
| Actions rapides | Reboot, stop, start depuis la TUI | Basse |
| Raccourcis numériques | `1-5` pour changer de produit | Basse |
| Breadcrumb | Fil d'Ariane en haut | Basse |

---

## 📝 Documentation générée

La commande `make doc` génère automatiquement la documentation. Le fichier `doc/ovhcloud_browser.md` devrait ressembler à :

```markdown
## ovhcloud browser

Launch a TUI simulating the OVHcloud Manager interface

### Synopsis

Launch an interactive Terminal User Interface that simulates the 
OVHcloud Manager (https://manager.eu.ovhcloud.com/).

Navigate through your Public Cloud services using keyboard controls.

### Usage

ovhcloud browser [flags]

### Keyboard shortcuts

Project Selection:
  ↑↓        Navigate through projects
  Enter     Select project and view its products
  d         Set selected project as default
  q         Quit

Product Navigation:
  ←→        Switch between products (Instances, K8s, DB, Storage, Networks)
  ↑↓        Navigate through list
  Enter     View resource details
  p         Return to project selection
  q         Quit

Detail View:
  Esc       Return to list
  p         Return to project selection
  q         Quit

### SEE ALSO

* [ovhcloud](ovhcloud.md) - CLI to manage your OVHcloud services
```

---

## 🔧 Commandes de développement

```bash
# Compiler
make build

# Tester
go test ./internal/services/browser/... -v

# Générer la doc
make doc

# Lancer le browser
./ovhcloud browser
```

---

## 📚 Dépendances

- **[Cobra](https://github.com/spf13/cobra)** : Framework CLI
- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** : Framework TUI
- **[Bubbles](https://github.com/charmbracelet/bubbles)** : Composants TUI (table)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** : Styling terminal
