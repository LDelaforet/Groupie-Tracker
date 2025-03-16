document.addEventListener('DOMContentLoaded', function() {
    // Sélectionner à la fois les étoiles de la page de résultats et les boutons favoris de la page de détail
    const favStars = document.querySelectorAll('.isFav');
    const favButtons = document.querySelectorAll('.favorite-button');
    
    console.log('Nombre d\'étoiles trouvées:', favStars.length);
    console.log('Nombre de boutons favoris trouvés:', favButtons.length);
    
    // Vérifier si nous sommes sur la page des favoris
    const isFavoritesPage = window.location.pathname === '/favorites';
    
    // Charger l'état initial des favoris
    fetch('/getFavorites')
        .then(response => response.json())
        .then(favorites => {
            console.log('Favoris reçus:', favorites);
            
            // Mettre à jour l'état des étoiles sur la page de résultats
            favStars.forEach(star => {
                const releaseId = star.getAttribute('data-release-id');
                const container = star.closest('.resultDiv');
                console.log('Vérification de l\'étoile:', releaseId, 'État:', favorites[releaseId]);
                
                if (favorites[releaseId] && favorites[releaseId].IsFavorite) {
                    star.src = "/pictures/star_filled.svg";
                } else {
                    star.src = "/pictures/star_empty.svg";
                    // Masquer les éléments non favoris si nous sommes sur la page des favoris
                    if (isFavoritesPage) {
                        container.style.display = 'none';
                    }
                }
            });
            
            // Mettre à jour l'état des boutons favoris sur la page de détail
            favButtons.forEach(button => {
                const releaseId = button.getAttribute('data-release-id');
                const starImg = button.querySelector('img');
                
                if (favorites[releaseId] && favorites[releaseId].IsFavorite) {
                    starImg.src = "/pictures/star_filled.svg";
                } else {
                    starImg.src = "/pictures/star_empty.svg";
                }
            });
        })
        .catch(error => console.error('Erreur lors du chargement des favoris:', error));
    
    // Gérer les clics sur les étoiles de la page de résultats
    favStars.forEach(star => {
        star.addEventListener('click', function(e) {
            e.preventDefault();
            e.stopPropagation();
            const container = this.closest('.resultDiv');
            const releaseId = this.getAttribute('data-release-id');
            
            // Récupérer toutes les informations de la release
            const title = container.querySelector('.resultName').textContent;
            const type = this.getAttribute('data-type');
            const typeDisp = container.querySelector('.resultTypeDisp').textContent;
            const year = container.querySelector('.resultYear').textContent;
            const country = container.querySelector('.resultCountry').textContent;
            const imageLink = container.querySelector('.resultImage').src;

            console.log('Clic sur l\'étoile:', releaseId, title);
            
            toggleFavorite(releaseId, title, imageLink, country, year, type, typeDisp, this);
        });
    });
    
    // Gérer les clics sur les boutons favoris de la page de détail
    favButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            e.stopPropagation();
            
            const releaseId = this.getAttribute('data-release-id');
            const type = this.getAttribute('data-type');
            const starImg = this.querySelector('img');
            
            // Pour la page de détail, nous devons extraire les informations différemment
            const title = document.querySelector('.release-title').textContent;
            const year = document.querySelector('.detail-item:nth-child(1) .detail-value').textContent;
            const country = document.querySelector('.detail-item:nth-child(2) .detail-value').textContent;
            const imageLink = document.querySelector('.release-cover img').src;
            const typeDisp = type; // Utiliser le type comme typeDisp par défaut
            
            console.log('Clic sur le bouton favori:', releaseId, title);
            
            toggleFavorite(releaseId, title, imageLink, country, year, type, typeDisp, starImg);
        });
    });
    
    // Fonction commune pour basculer l'état favori
    function toggleFavorite(releaseId, title, imageLink, country, year, type, typeDisp, element) {
        fetch('/toggleFavorite', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                releaseId: releaseId.toString(),
                title: title,
                imageLink: imageLink,
                country: country,
                year: year,
                type: type,
                typeDisp: typeDisp
            })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            console.log('Réponse reçue:', data);
            if (data.IsFavorite) {
                element.src = "/pictures/star_filled.svg";
                if (isFavoritesPage && element.closest('.resultDiv')) {
                    element.closest('.resultDiv').style.display = 'block';
                }
            } else {
                element.src = "/pictures/star_empty.svg";
                if (isFavoritesPage && element.closest('.resultDiv')) {
                    element.closest('.resultDiv').style.display = 'none';
                }
            }
        })
        .catch(error => {
            console.error('Erreur:', error);
        });
    }

    const images = document.querySelectorAll('img');
    images.forEach(img => {
        if (img.src === 'https://st.discogs.com/da98ce6cf72bc2cc9cc9603a86329ebb3ed4b3fb/images/spacer.gif') {
            img.src = '/pictures/noPicture.png';
        }
    });
}); 