# Identifiants Discogs pour le rendu de tp:
user: LDelaforet
pass: "^&g],yjJ=43Nqv


# Discogs - Groupie-Tracker

Ce projet a été réalisé par **Léo Velazquez** dans le cadre d'un projet à **Ynov**.

---

## 🚀 Accès au site

### Démarrer le projet

1. Executez les commandes suivantes:
    cd src
    go run main.go
2. Ouvrez votre navigateur et entrez l'adresse suivante :
   ```
   http://localhost:8080
   ```

### API utilisée

Groupie-Tracker exploite l'API [Discogs](https://www.discogs.com/developers).

---

## 🛠 Développement

### Choix de l'API

Plusieurs API me sont venues en tete au moment du developpement mais celle de discogs a retenue mon attention de part sa facilité d'utilisation ainsi que sa polyvalence.

### Processus de développement

1. Développement du **back-end**.
2. Mise en place du **front-end**.
3. Ajout de fonctionnalitées annexes.

### Organisation du projet

- **Décomposition du projet** : J'ai séparé le projet en différentes phases, allant de la recherche API à l'implémentation des fonctionnalités.
- **Gestion des tâches** : Travaillant seul, j'ai planifié et priorisé chaque étape du développement en suivant une approche itérative.
- **Documentation** : Je me suis appuyé sur des ressources en ligne, la documentation de l'API et des tutoriels pour garantir un projet bien structuré et fonctionnel.

---

## 🔗 Endpoints utilisés

- `https://api.discogs.com/artists/{id}` → Récupération des informations d'un artiste
- `https://api.discogs.com/releases/{id}` → Détails d'une sortie musicale
- `https://api.discogs.com/masters/{id}` → Informations sur un enregistrement maître
- `https://api.discogs.com/users/{username}` → Informations sur un utilisateur Discogs (utilisée pour verifier l'authentification)

---

## 🌐 Routes du site

### Pages principales

- **Accueil** : `http://localhost:8080/`
- **Recherche d'artiste** : `http://localhost:8080/searchArtist`
- **Détails de l'artiste** : `http://localhost:8080/artistDetails`
- **Détails de la sortie musicale** : `http://localhost:8080/releaseDetails`
- **Voir les favoris** : `http://localhost:8080/favorites`