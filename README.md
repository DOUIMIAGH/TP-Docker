# tp-docker

## Exercice 1 — Comprendre le cycle de vie des processus dans un conteneur

### Question 1 — Rôle du processus principal

Le processus principal exécute la commande prévue pour le conteneur. Dans notre essai, `sleep 300` a maintenu le conteneur actif. Quand ce processus s’arrête, le conteneur s’arrête également.

### Question 2 — Différence entre `docker stop` et `docker kill`

`docker stop` envoie d’abord `SIGTERM` au processus principal et lui laisse le délai prévu pour s’arrêter ; si le délai expire, Docker peut envoyer `SIGKILL`. `docker kill` envoie directement le signal choisi, `SIGKILL` par défaut. Dans notre essai, `--signal=TERM` demandait explicitement à `docker kill` d’envoyer `SIGTERM`.

### Question 3 — Pourquoi le conteneur s’arrête lorsque son processus principal se termine

Docker associe l’état du conteneur à son processus principal. Quand celui-ci se termine, le conteneur passe à l’état arrêté. Dans notre essai, le processus principal était `sleep 300` ; après l’envoi de `SIGTERM`, son état est devenu `exited`.

### Question 4 — Différence entre `SIGTERM` et `SIGKILL`

`SIGTERM` demande au processus de s’arrêter proprement et lui laisse la possibilité de fermer ses ressources. `SIGKILL` force son arrêt immédiat ; le processus ne peut ni le traiter ni l’ignorer. C’est pourquoi `docker stop` commence par `SIGTERM`, puis peut envoyer `SIGKILL` si le délai d’arrêt expire. Notre commande `docker kill --signal=TERM` avait explicitement choisi `SIGTERM`.

### Question 5 — Pourquoi une application doit gérer les signaux d’arrêt

En gérant `SIGTERM`, une application peut terminer ses opérations en cours, fermer ses connexions et libérer ses ressources avant de s’arrêter. Elle réduit ainsi le risque d’interrompre un traitement ou de laisser des données dans un état incohérent lorsque Docker lui demande de s’arrêter.

## Exercice 2 — Comprendre les couches d’une image et le système de fichiers

### Question 1 — Différence entre les couches d’une image et la couche inscriptible d’un conteneur

Les couches d’une image forment son système de fichiers en lecture seule. Docker ajoute une couche inscriptible propre à chaque conteneur pour y enregistrer ses changements. Notre `docker diff` a signalé les fichiers créés ou modifiés dans `couches-test`, tandis que l’image Alpine restait inchangée.

### Question 2 — Rôle du mécanisme *copy-on-write*

Le mécanisme *copy-on-write* permet au conteneur de réutiliser les couches de l’image sans les modifier. Lorsqu’un fichier partagé doit être modifié, Docker place une copie dans la couche inscriptible du conteneur, puis applique le changement à cette copie. L’image d’origine reste donc intacte.

### Question 3 — Pourquoi les modifications d’un conteneur ne changent pas l’image d’origine

L’image est composée de couches en lecture seule. Les écritures du conteneur sont enregistrées dans sa couche inscriptible, qui est séparée de ces couches. La modification de `/etc/alpine-release` observée avec `docker diff` appartenait donc à `couches-test`, pas à l’image `alpine:3.22`.

### Question 4 — Rôle de `docker diff`

`docker diff` affiche les ajouts, les modifications et les suppressions effectués dans le système de fichiers d’un conteneur par rapport à son image de départ. Les indicateurs sont `A` pour un ajout, `C` pour un changement et `D` pour une suppression. Pour `couches-test`, la commande a notamment montré `A /tmp/test.txt`, `A /opt/test` et `C /etc/alpine-release`.

### Question 5 — Pourquoi les modifications de la couche inscriptible disparaissent avec le conteneur

La couche inscriptible appartient au conteneur et est supprimée avec lui. Après avoir supprimé `couches-test`, nous avons créé un nouveau conteneur depuis `alpine:3.22` : `/etc/alpine-release` affichait seulement `3.22.7`, sans notre modification. L’image d’origine n’avait donc pas été changée.

## Exercice 3 — Construire une image avec plusieurs étapes (Multi-stage build)

### Question 1 — Pourquoi utiliser plusieurs instructions `FROM` ?

Chaque instruction `FROM` ouvre une étape distincte. Cela permet d'utiliser une image avec le compilateur pour construire l'application, puis de produire l'image d'exécution avec seulement les éléments nécessaires. Ici, l'étape `builder` compile le programme Go et l'étape finale contient l'exécutable.

### Question 2 — Différence entre l'étape `builder` et l'étape finale

L'étape `builder` utilise `golang:1.24-alpine`, contient le code source et compile l'application. L'étape finale part de `scratch`, une image vide, et reçoit le binaire compilé. Nous avons vérifié que ce binaire démarre le serveur et répond à la requête HTTP attendue.

### Question 3 — Pourquoi l'image finale ne contient-elle pas le compilateur Go ?

Le compilateur Go est utilisé uniquement dans l'étape `builder`. L'instruction `COPY --from=builder /app /app` copie seulement l'exécutable dans l'étape finale. Le compilateur et le code source ne sont donc pas inclus dans l'image finale.

### Question 4 — Avantages d'une image finale minimale

Une image plus petite demande moins de données à transférer et contient moins d'outils et de composants inutiles à l'exécution. L'image `application-go:v1` construite pour cet exercice a une taille de 8,15 Mo.

### Question 5 — Précautions avant d'utiliser une image `scratch` en production

Il faut vérifier que l'exécutable fonctionne dans une image vide et inclure explicitement les fichiers ou dépendances nécessaires à l'application. Il faut aussi tenir compte de l'absence de shell et d'outils système, qui rend les vérifications et le diagnostic plus limités. Dans cet exercice, le binaire a été testé en répondant à la requête HTTP prévue.

## Exercice 4 — Renforcer la sécurité d'un conteneur

### Question 1 — Pourquoi utiliser un utilisateur non privilégié ?

Un utilisateur non privilégié dispose de droits plus limités qu'un utilisateur root. Si l'application rencontre un problème ou est compromise, les actions possibles dans le conteneur sont ainsi limitées. Nos commandes ont confirmé que le processus s'exécutait avec `uid=10001` et qu'il ne pouvait pas créer un fichier dans `/root`.

### Question 2 — Que sont les capacités Linux ?

Les capacités Linux sont des privilèges précis que le noyau peut accorder à un processus. Elles permettent de limiter les actions privilégiées disponibles. L'option `--cap-drop=ALL` utilisée dans le test retire les capacités du conteneur.

### Question 3 — Différence entre `--read-only` et `--tmpfs`

`--read-only` rend le système de fichiers racine du conteneur non modifiable. `--tmpfs` fournit un emplacement temporaire inscriptible en mémoire. Dans notre test, la racine était en lecture seule, tandis que la création d'un fichier dans `/tmp` a réussi.

### Question 4 — Pourquoi conserver des zones inscriptibles ?

Une application peut avoir besoin d'écrire des fichiers temporaires ou des données nécessaires à son fonctionnement. On peut lui fournir un emplacement limité comme `/tmp` avec `--tmpfs`, sans rendre inscriptible tout le système de fichiers racine.

### Question 5 — Ces options garantissent-elles une isolation parfaite ?

Non. Elles réduisent les privilèges du processus et limitent les écritures, mais ne suffisent pas à garantir une isolation parfaite. Un conteneur partage notamment le noyau de l'hôte ; ces options font donc partie des mesures de sécurité à appliquer, sans remplacer une configuration et des protections adaptées.

## Exercice 5 — Diagnostiquer un conteneur qui se termine anormalement

### Question 1 — Différence entre un conteneur arrêté et un conteneur supprimé

Un conteneur arrêté existe toujours dans Docker : on peut encore consulter son état, ses métadonnées et ses journaux. Dans notre test, `diagnostic-test` apparaissait comme `Exited (42)` avant sa suppression. Après `docker rm`, il n'apparaissait plus dans `docker ps -a`.

### Question 2 — Signification d'un code de sortie différent de zéro

Un code de sortie différent de zéro indique que le processus s'est terminé avec un résultat signalant un échec. Dans cet exercice, la commande terminait volontairement avec le code `42`, que Docker a bien enregistré.

### Question 3 — Différence entre la sortie standard et la sortie d'erreur

La sortie standard sert aux messages et résultats habituels du programme. La sortie d'erreur sert aux messages de diagnostic ou d'échec. Notre commande a écrit `Demarrage de l application` sur la sortie standard et `Erreur simulee` sur la sortie d'erreur ; `docker logs` a affiché les deux messages.

### Question 4 — Utilité des journaux quand un conteneur se termine rapidement

Les journaux conservent les messages produits par le processus et permettent de comprendre ce qui s'est passé après son arrêt. Ici, même si le conteneur s'est terminé immédiatement, `docker logs` a montré le message de démarrage et l'erreur simulée.

### Question 5 — Pourquoi examiner le code de sortie

Le code de sortie donne un résultat vérifiable du processus, même si son message est absent ou peu explicite. Dans notre test, l'état `Exited (42)` et `docker inspect` confirmaient l'échec volontaire avec le code attendu, en complément des messages des journaux.

## Exercice 6 — Configurer la rotation des journaux

### Question 1 — Pourquoi des journaux trop volumineux posent-ils problème ?

Des journaux qui grossissent sans limite peuvent occuper l'espace disque du serveur et gêner son fonctionnement. Ils peuvent aussi rendre la consultation des messages utiles plus difficile. Dans notre test, le pilote `local` était configuré avec une taille cible de `1m` et un maximum de deux fichiers.

### Question 2 — Rôle d'un pilote de journaux Docker

Le pilote définit comment Docker collecte et stocke les journaux d'un conteneur. Nous avons observé le pilote `local` pour `logs-test` et le pilote `json-file` pour `bigdata-s01-outils-1`. La commande `docker logs` nous a permis de consulter les messages de `logs-test`.

### Question 3 — Différence entre `docker logs` et le stockage à long terme

`docker logs` sert à consulter les messages capturés pour un conteneur. Cette commande ne définit pas à elle seule leur durée de conservation. Celle-ci dépend de la configuration du pilote et de la stratégie de stockage choisie pour conserver les journaux dans le temps.

### Question 4 — Pourquoi la rotation est-elle importante sur un serveur ?

La rotation limite la croissance des fichiers de journaux en conservant un nombre et une taille définis. Cela aide à éviter que les journaux consomment tout l'espace disque du serveur lorsque les applications produisent beaucoup de messages.

### Question 5 — Pourquoi adapter la stratégie à la production ?

Le volume de messages, l'espace disponible et la durée de conservation nécessaire varient selon l'application et l'environnement. Il faut donc choisir le pilote et les limites en fonction de ces besoins. Dans notre comparaison, `logs-test` avait des limites explicites, tandis que `bigdata-s01-outils-1` utilisait `json-file` sans options de rotation configurées au niveau du conteneur.

## Exercice 7 — Explorer les processus et les modifications d'un conteneur en fonctionnement

### Question 1 — Différence entre `docker exec` et `docker run`

`docker run` crée un nouveau conteneur à partir d'une image et lance son processus principal. `docker exec` lance une commande supplémentaire dans un conteneur déjà actif. Dans notre test, `docker exec ... ps` a ajouté le processus `ps` sans remplacer le processus principal `sh`.

### Question 2 — Rôle de `docker top`

`docker top` affiche les processus en cours dans un conteneur. Pour `inspection-test`, la commande a montré le processus principal `sh` et son processus enfant `sleep 30`.

### Question 3 — Utilité de `docker diff` pendant un diagnostic

`docker diff` indique les fichiers ajoutés, modifiés ou supprimés par rapport à l'image de départ. Il permet de repérer les changements faits dans le conteneur. Dans notre test, il a signalé l'ajout de `/tmp/diagnostic.txt`.

### Question 4 — Informations fournies par `docker stats`

`docker stats` affiche l'utilisation des ressources du conteneur, notamment le CPU, la mémoire, les entrées/sorties réseau et disque ainsi que le nombre de processus. La mesure ponctuelle de `inspection-test` indiquait `0,00 %` de CPU, `444 KiB` de mémoire et deux processus.

### Question 5 — Pourquoi documenter les modifications en production ?

Une modification manuelle peut être difficile à retrouver et à reproduire. La documenter permet de comprendre l'état du conteneur, de garder une trace des changements et de recréer une configuration cohérente lors d'un nouveau déploiement.

## Conclusion

Le build multi-stage a séparé la compilation de l'exécution et produit une image finale `scratch` de 8,15 Mo contenant l'application. La sécurité a été renforcée en utilisant un utilisateur non privilégié, un système de fichiers racine en lecture seule, un espace temporaire `tmpfs`, la suppression des capacités et l'option `no-new-privileges`.
