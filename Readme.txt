Utilisation :
  Lancer le fichier main.go puis se rendre sur l'adresse "localhost:8080" avec votre navigateur. Vous voila sur le site.

Les routes :
  Globalement, le header contient tous les liens (à part la recherche qui se fait forcément a l'index)
  "/", c'est l'index. D'ici, on peut se rendre sur la page de recherche.
  "/search", c'est la recherche, tapez votre recherche et sélectionnez le filtre "track", "album" ou "artist".
  "/result", montre les résultats de la recherche en pagination, 10 résultats par pages.
  "/connect", pour vous connecter à votre compte.
  "/connectHandler" sert a analyser les entrées de la route "/connect" et rendre la connection effective ou vous ramener vers "/connect" avec la bonne erreur.
  "/register", pour créer un compte.
  "/registerHandler" sert a analyser les entrées de la route "/register" et rendre la création effective ou vous ramener vers "/register" avec la bonne erreur.
  "/tracks" renvoie des tracks (la recherche de base sur "Dethklok" parcqu'on ne peut pas chercher de sons aléatoires avec l'API Deezer)
  "/albums" idem que tracks
  "/artists" idem que tracks
  "/track" renvoie les infos d'un track en particulier à partir de son id avec la possibilité de l'ajouter aux favoris.
  "/album" idem que track
  "/artist" idem que track
  "/favoris" vous montre vos sons/albums/artistes favoris séparément.
  "/AddToFavorite{type}" ajoute le type (track/album/artiste) dans la liste de favoris correspondante.
