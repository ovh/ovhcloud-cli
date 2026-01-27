# Spécification : Browser TUI - Module Instances

## Vue d'ensemble

Le module Instances du Browser TUI permet de gérer les instances cloud (VMs) OVHcloud de manière interactive. Il offre une navigation complète, la création d'instances via un wizard, et des actions rapides sur les instances existantes.

## Architecture

### Fichiers principaux

| Fichier | Description |
|---------|-------------|
| `internal/services/browser/manager.go` | Gestion de l'état, rendu des vues, gestion des touches |
| `internal/services/browser/api.go` | Appels API OVHcloud, logique métier asynchrone |

### Types de vues (ViewMode)

```go
const (
    LoadingView                       // Chargement en cours
    ProjectSelectView                 // Sélection du projet cloud
    TableView                         // Liste des instances
    EmptyView                         // Aucune instance
    DetailView                        // Détail d'une instance
    DeleteConfirmView                 // Confirmation de suppression
    WizardView                        // Wizard de création
)
```

### Structure WizardData

```go
type WizardData struct {
    step                       WizardStep
    regions                    []map[string]interface{}
    flavors                    []map[string]interface{}
    images                     []map[string]interface{}
    sshKeys                    []map[string]interface{}
    privateNetworks            []map[string]interface{}
    floatingIPs                []map[string]interface{}
    
    // Sélections
    selectedRegion             string
    selectedFlavor             string
    selectedImage              string
    selectedSSHKey             string
    selectedPrivateNetwork     string
    selectedSubnetId           string
    selectedFloatingIP         string
    usePublicNetwork           bool
    instanceName               string
    
    // Création de réseau inline
    creatingNetwork            bool
    newNetworkName             string
    newNetworkVlanId           int
    newNetworkCIDR             string
    newNetworkDHCP             bool
    
    // États
    isLoading                  bool
    loadingMessage             string
    errorMsg                   string
    filterMode                 bool
    filterInput                string
    
    // Tracking pour cleanup
    createdNetworkId           string
    createdInstanceId          string
    createdGatewayId           string
    createdFloatingIPId        string
    cleanupPending             bool
    cleanupError               string
}
```

---

## Fonctionnalités implémentées

### 1. Liste des instances (TableView)

**Colonnes affichées :**
- Nom
- Status (avec icône couleur)
- Région
- IP (IPv4 publique)
- Flavor

**Actions disponibles :**
| Touche | Action |
|--------|--------|
| `↑/↓` | Naviguer dans la liste |
| `Enter` | Voir les détails |
| `/` | Filtrer par nom |
| `c` | Créer une instance |
| `Del` | Supprimer une instance |
| `←/→` | Changer de produit |
| `p` | Changer de projet |
| `q` | Quitter |

### 2. Détail d'une instance (DetailView)

**Informations affichées :**
- Status (avec indicateur visuel 🟢/🔴)
- ID
- Région
- Flavor
- Image
- Date de création
- IPv4 / IPv6

**Actions rapides :**
| Index | Action | Description |
|-------|--------|-------------|
| 0 | Reboot | Redémarrage soft |
| 1 | Rescue Mode / Exit Rescue | Mode rescue (toggle) |
| 2 | Stop / Start | Arrêt ou démarrage (selon état) |
| 3 | Console | Obtenir URL VNC |
| 4 | Reinstall | Réinstaller avec l'image actuelle |

**Navigation :**
| Touche | Action |
|--------|--------|
| `←/→` | Naviguer entre les actions |
| `Enter` | Exécuter l'action (avec confirmation) |
| `Esc` | Retour à la liste |

### 3. Suppression d'instance (DeleteConfirmView)

- L'utilisateur doit taper le nom exact de l'instance pour confirmer
- Endpoint : `DELETE /v1/cloud/project/{project}/instance/{id}`

---

## Wizard de création d'instance

### Étapes du wizard (WizardStep)

```go
const (
    WizardStepRegion     // Sélection de la région
    WizardStepFlavor     // Sélection du flavor (taille VM)
    WizardStepImage      // Sélection de l'image OS
    WizardStepSSHKey     // Sélection de la clé SSH
    WizardStepNetwork    // Configuration réseau
    WizardStepFloatingIP // Sélection Floating IP (si réseau privé seul)
    WizardStepName       // Nom de l'instance
    WizardStepConfirm    // Confirmation finale
)
```

### Flux de navigation

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│   Region ──► Flavor ──► Image ──► SSH Key ──► Network ──► Name     │
│                                                      │              │
│                                                      ▼              │
│                                               [Réseau privé seul?]  │
│                                                      │              │
│                                              Oui ◄───┴───► Non      │
│                                               │              │      │
│                                               ▼              │      │
│                                         Floating IP          │      │
│                                               │              │      │
│                                               ▼              ▼      │
│                                             Name ◄──────────┘      │
│                                               │                     │
│                                               ▼                     │
│                                           Confirm                   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Détail de chaque étape

#### Étape 1 : Région (WizardStepRegion)
- **API** : `GET /v1/cloud/project/{project}/region`
- **Affichage** : Liste des régions disponibles
- **Filtrage** : `/` pour filtrer par nom
- **Message loading** : "Loading regions..."

#### Étape 2 : Flavor (WizardStepFlavor)
- **API** : `GET /v1/cloud/project/{project}/flavor?region={region}`
- **Affichage** : Nom, vCPUs, RAM, catégorie
- **Filtrage** : `/` pour filtrer
- **Message loading** : "Loading flavors..."

#### Étape 3 : Image (WizardStepImage)
- **API** : `GET /v1/cloud/project/{project}/image?region={region}`
- **Affichage** : Nom, type (Linux/Windows), taille
- **Filtrage** : `/` pour filtrer
- **Message loading** : "Loading images..."

#### Étape 4 : Clé SSH (WizardStepSSHKey)
- **API** : `GET /v1/cloud/project/{project}/sshkey`
- **Options** :
  - "(No SSH Key)" - pas de clé
  - Liste des clés existantes
- **Filtrage** : `/` pour filtrer
- **Message loading** : "Loading SSH keys..."

#### Étape 5 : Réseau (WizardStepNetwork)
- **API** : `GET /v1/cloud/project/{project}/network/private`
- **Options** :
  - Toggle "Public Network" (activé par défaut)
  - "(No Private Network)"
  - "+ Create new private network" → sous-wizard
  - Liste des réseaux privés existants
- **Filtrage** : `/` pour filtrer les réseaux
- **Message loading** : "Loading networks..."

##### Création de réseau inline
Si l'utilisateur choisit "Create new private network" :
- Champs : Nom, VLAN ID (1-4094), CIDR, DHCP (toggle)
- Navigation : `↑/↓` entre champs, `Tab`/`Enter` pour avancer
- **Messages loading** :
  - "Creating network..."
  - "Creating subnet..."

#### Étape 6 : Floating IP (WizardStepFloatingIP) - Conditionnelle
**Condition** : Réseau privé sélectionné ET réseau public désactivé

- **API** : `GET /v1/cloud/project/{project}/region/{region}/floatingip`
- **Options** :
  - "(No Floating IP - no external access)"
  - "+ Create new Floating IP"
  - Liste des floating IPs disponibles (non associées)
- **Message loading** : "Loading floating IPs..."

#### Étape 7 : Nom (WizardStepName)
- Champ texte libre
- Validation : non vide

#### Étape 8 : Confirmation (WizardStepConfirm)
- Récapitulatif de toutes les sélections
- Boutons : `[Create]` / `[Cancel]`

---

## Processus de création d'instance

### Cas 1 : Réseau public (standard)

```
POST /v1/cloud/project/{project}/instance
Body: {
    "flavorId": "...",
    "imageId": "...",
    "name": "...",
    "region": "...",
    "sshKeyId": "..." (optionnel)
}
```
- **Message loading** : "Creating instance..."

### Cas 2 : Réseau privé + Floating IP

```
1. POST /v1/cloud/project/{project}/instance
   Body: {
       "flavorId": "...",
       "imageId": "...",
       "name": "...",
       "region": "...",
       "networks": [{"networkId": "..."}]
   }
   Message: "Creating instance..."

2. Polling GET /v1/cloud/project/{project}/instance/{id}
   Attente que l'instance ait une IP privée
   Message: "Waiting for instance private IP..."
   (polling toutes les 5s, max 60s)

3. POST /v1/cloud/project/{project}/region/{region}/instance/{id}/floatingIp
   Body: {
       "ip": "<private_ip>",
       "gateway": {"model": "s", "name": "gw-<instance_name>"}
   }
   Message: "Creating gateway and attaching Floating IP..."
```

### Gestion d'erreur et cleanup

Si une erreur survient après la création de ressources :
1. Affichage d'une boîte de dialogue de confirmation
2. Liste des ressources créées (instance, réseau, gateway, floating IP)
3. Options : "Yes, delete all" / "No, keep them"
4. **Message loading** : "Cleaning up resources..."

**Endpoints de cleanup** :
- Instance : `DELETE /v1/cloud/project/{project}/instance/{id}`
- Network : `DELETE /v1/cloud/project/{project}/network/private/{id}`
- Floating IP : (via instance delete, cascade)

---

## Messages de progression (loadingMessage)

| Opération | Message |
|-----------|---------|
| Démarrage wizard | "Loading regions..." |
| Sélection région | "Loading flavors..." |
| Sélection flavor | "Loading images..." |
| Sélection image | "Loading SSH keys..." |
| Sélection SSH key | "Loading networks..." |
| Sélection réseau privé | "Loading floating IPs..." |
| Création réseau | "Creating network..." |
| Création subnet | "Creating subnet..." |
| Création instance | "Creating instance..." |
| Attente IP | "Waiting for instance private IP..." |
| Floating IP | "Creating gateway and attaching Floating IP..." |
| Cleanup | "Cleaning up resources..." |

---

## Types de messages (tea.Msg)

### Messages de données
```go
type regionsLoadedMsg struct { regions []map[string]interface{}; err error }
type flavorsLoadedMsg struct { flavors []map[string]interface{}; err error }
type imagesLoadedMsg struct { images []map[string]interface{}; err error }
type sshKeysLoadedMsg struct { sshKeys []map[string]interface{}; err error }
type privateNetworksLoadedMsg struct { networks []map[string]interface{}; err error }
type floatingIPsLoadedMsg struct { floatingIPs []map[string]interface{}; err error }
```

### Messages de création
```go
type instanceCreatedMsg struct { instance map[string]interface{}; err error }
type networkCreatedMsg struct { network map[string]interface{}; err error }
type networkStepMsg struct { step string; networkId string; network map[string]interface{}; err error }
type instanceIPReadyMsg struct { instanceId, instanceName, privateIP string; err error }
type floatingIPAttachedMsg struct { instanceName string; err error }
```

### Messages d'action
```go
type instanceActionMsg struct { action, instanceId string; err error }
type instanceDeletedMsg struct { success bool; instanceId string; err error }
type cleanupCompletedMsg struct { deletedResources, errors []string }
type progressMsg struct { message string }
```

---

## Ce qui reste à faire (TODO)

### Priorité haute

1. **Shelve/Unshelve** : Mettre en veille prolongée une instance
   - `POST /v1/cloud/project/{project}/instance/{id}/shelve`
   - `POST /v1/cloud/project/{project}/instance/{id}/unshelve`

2. **Resize** : Changer le flavor d'une instance
   - `POST /v1/cloud/project/{project}/instance/{id}/resize`
   - Wizard pour sélectionner le nouveau flavor

3. **Snapshot** : Créer une image de l'instance
   - `POST /v1/cloud/project/{project}/instance/{id}/snapshot`
   - Champ pour le nom du snapshot

4. **Attach Volume** : Attacher un volume block storage
   - `POST /v1/cloud/project/{project}/instance/{id}/attachVolume`
   - Liste des volumes disponibles

### Priorité moyenne

5. **Modifier le billing** : Monthly ↔ Hourly
   - `POST /v1/cloud/project/{project}/instance/{id}/activeMonthlyBilling`

6. **Backup automatique** : Configurer les backups
   - Afficher les backups existants
   - Créer/supprimer des schedules

7. **Security Groups** : Gérer les règles de firewall
   - Lister les security groups
   - Associer/dissocier

8. **Private networks** : Attacher/détacher des réseaux
   - `POST /v1/cloud/project/{project}/instance/{id}/interface`

### Priorité basse

9. **Logs/Monitoring** : Afficher les logs de l'instance

10. **Tags/Labels** : Ajouter des métadonnées

11. **Bulk actions** : Actions sur plusieurs instances

---

## Points d'attention pour le développement

### Gestion des états asynchrones

Les opérations API sont asynchrones (tea.Cmd). Chaque opération doit :
1. Mettre `isLoading = true` et `loadingMessage = "..."`
2. Retourner une `tea.Cmd` qui fait l'appel API
3. Le handler du message reçu met `isLoading = false` et `loadingMessage = ""`

### Pattern de message step-by-step

Pour les opérations multi-étapes (ex: création réseau + subnet) :
```go
// Étape 1 retourne un message intermédiaire
return networkStepMsg{step: "network_created", networkId: id}

// Handler déclenche l'étape 2
case "network_created":
    return m, m.createSubnet(msg.networkId)
```

### Filtrage des listes

Utiliser `filterMode` et `filterInput` pour le filtrage :
```go
if m.wizard.filterMode {
    // Capturer l'input texte
    m.wizard.filterInput += msg.String()
}
// Filtrer la liste
filtered := filterByName(items, m.wizard.filterInput)
```

### Navigation backward

Chaque étape doit pouvoir revenir en arrière (`←` ou `backspace`) :
```go
case "left", "backspace":
    m.wizard.step = previousStep
    m.wizard.filterInput = ""
```

### Cleanup des ressources

Tracker les IDs des ressources créées :
```go
m.wizard.createdNetworkId = networkId
m.wizard.createdInstanceId = instanceId
```

Vérifier avant cleanup :
```go
func (m Model) hasCreatedResources() bool {
    return m.wizard.createdNetworkId != "" || 
           m.wizard.createdInstanceId != "" || ...
}
```

---

## Endpoints API utilisés

### Lecture
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/v1/cloud/project/{project}/instance` | Liste des instances |
| GET | `/v1/cloud/project/{project}/instance/{id}` | Détail d'une instance |
| GET | `/v1/cloud/project/{project}/region` | Liste des régions |
| GET | `/v1/cloud/project/{project}/flavor?region={r}` | Flavors par région |
| GET | `/v1/cloud/project/{project}/image?region={r}` | Images par région |
| GET | `/v1/cloud/project/{project}/sshkey` | Clés SSH |
| GET | `/v1/cloud/project/{project}/network/private` | Réseaux privés |
| GET | `/v1/cloud/project/{project}/network/private/{id}/subnet` | Subnets |
| GET | `/v1/cloud/project/{project}/region/{r}/floatingip` | Floating IPs |

### Création
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/v1/cloud/project/{project}/instance` | Créer instance |
| POST | `/v1/cloud/project/{project}/network/private` | Créer réseau |
| POST | `/v1/cloud/project/{project}/network/private/{id}/subnet` | Créer subnet |
| POST | `/v1/cloud/project/{project}/region/{r}/instance/{id}/floatingIp` | Attacher floating IP |

### Actions
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/v1/cloud/project/{project}/instance/{id}/reboot` | Redémarrer |
| POST | `/v1/cloud/project/{project}/instance/{id}/start` | Démarrer |
| POST | `/v1/cloud/project/{project}/instance/{id}/stop` | Arrêter |
| POST | `/v1/cloud/project/{project}/instance/{id}/rescueMode` | Mode rescue |
| POST | `/v1/cloud/project/{project}/instance/{id}/reinstall` | Réinstaller |
| POST | `/v1/cloud/project/{project}/instance/{id}/vnc` | Console VNC |
| DELETE | `/v1/cloud/project/{project}/instance/{id}` | Supprimer |

### Cleanup
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| DELETE | `/v1/cloud/project/{project}/instance/{id}` | Supprimer instance |
| DELETE | `/v1/cloud/project/{project}/network/private/{id}` | Supprimer réseau |

---

## Exemple : Ajouter une nouvelle action

Pour ajouter l'action "Shelve" :

1. **api.go** - Ajouter dans `executeInstanceAction` :
```go
actions := []string{"reboot", "rescue", "stop_or_start", "vnc", "reinstall", "shelve"}

case "shelve":
    status := strings.ToUpper(getString(m.detailData, "status"))
    if status == "SHELVED" || status == "SHELVED_OFFLOADED" {
        action = "unshelve"
        endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/unshelve", m.cloudProject, instanceId)
        err = httpLib.Client.Post(endpoint, nil, nil)
    } else {
        endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/shelve", m.cloudProject, instanceId)
        err = httpLib.Client.Post(endpoint, nil, nil)
    }
```

2. **api.go** - Ajouter dans `handleInstanceAction` :
```go
actionNames := map[string]string{
    // ...existants...
    "shelve":   "Shelve",
    "unshelve": "Unshelve",
}
```

3. **manager.go** - Ajouter dans `renderInstanceDetail` :
```go
shelveAction := "Shelve"
if strings.ToUpper(status) == "SHELVED" || strings.ToUpper(status) == "SHELVED_OFFLOADED" {
    shelveAction = "Unshelve"
}
actions := []string{"Reboot", rescueAction, stopStartAction, "Console", "Reinstall", shelveAction}
```

---

## Tests recommandés

### Scénarios de création
1. Instance avec réseau public uniquement
2. Instance avec réseau privé + public
3. Instance avec réseau privé seul + nouveau floating IP
4. Instance avec réseau privé seul + floating IP existant
5. Création de réseau inline pendant le wizard
6. Annulation à chaque étape
7. Erreur API et cleanup

### Scénarios d'actions
1. Reboot d'une instance ACTIVE
2. Stop/Start cycle
3. Rescue mode entry/exit
4. Console VNC
5. Reinstall
6. Suppression avec confirmation

### Scénarios de navigation
1. Filtrage dans chaque liste
2. Navigation backward à chaque étape
3. Changement de projet pendant le wizard
