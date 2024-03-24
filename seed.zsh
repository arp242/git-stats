#!/usr/bin/env zsh

set -eu

# List of popular repos from https://gitstar-ranking.com, and some other random
# stuff I thought of. Sorted in time it takes to import on my laptop.
repos=(
	https://github.com/dominictarr/event-stream                   #    0s
	https://github.com/tukaani-project/xz                         #    0s
	https://git.savannah.gnu.org/git/gzip                         #    0s
	https://github.com/madler/zlib                                #    0s
	https://github.com/nvbn/thefuck                               #    0s
	git://git.kernel.org/pub/scm/linux/kernel/git/jberg/iw        #    0s
	git://git.kernel.org/pub/scm/utils/dash/dash                  #    0s
	git://git.sv.gnu.org/sed                                      #    0s
	https://github.com/xiph/ogg                                   #    0s
	https://gitlab.com/federicomenaquintero/bzip2                 #    0s
	https://gitlab.freedesktop.org/mesa/glu                       #    0s
	https://gitlab.freedesktop.org/wayland/wayland                #    0s
	https://github.com/necolas/normalize.css                      #    0s
	https://github.com/square/retrofit                            #    0s
	https://github.com/impress/impress.js                         #    0s
	https://github.com/iamkun/dayjs                               #    0s
	https://github.com/gulpjs/gulp                                #    0s
	https://github.com/BurntSushi/toml                            #    0s
	https://github.com/fsnotify/fsnotify                          #    1s
	https://github.com/ohmyzsh/ohmyzsh                            #    1s
	https://github.com/gwsw/less                                  #    1s
	https://git.savannah.gnu.org/git/screen                       #    1s
	https://github.com/NLnetLabs/ldns                             #    1s
	https://gitlab.freedesktop.org/fontconfig/fontconfig          #    1s
	https://gitlab.freedesktop.org/pixman/pixman                  #    1s
	https://github.com/xiph/vorbis                                #    1s
	https://github.com/jqlang/jq                                  #    1s
	https://github.com/laravel/laravel                            #    1s
	https://github.com/pallets/flask                              #    1s
	https://gitlab.gnome.org/GNOME/atk                            #    1s
	https://github.com/typicode/json-server                       #    1s
	https://github.com/expressjs/express                          #    1s
	https://github.com/httpie/httpie                              #    1s
	https://github.com/Genymobile/scrcpy                          #    1s
	https://github.com/gin-gonic/gin                              #    1s
	https://github.com/soimort/you-get                            #    1s
	https://github.com/juliangarnier/anime                        #    1s
	https://github.com/junegunn/fzf                               #    1s
	https://github.com/caddyserver/caddy                          #    1s
	https://github.com/PhilJay/MPAndroidChart                     #    1s
	https://github.com/AFNetworking/AFNetworking                  #    1s
	https://github.com/hexojs/hexo                                #    1s
	https://github.com/alvarotrigo/fullPage.js                    #    1s
	https://github.com/bailicangdu/vue2-elm                       #    1s
	https://github.com/htop-dev/htop                              #    1s
	https://github.com/socketio/socket.io                         #    2s
	https://github.com/ElemeFE/element                            #    2s
	https://github.com/fatedier/frp                               #    2s
	https://github.com/Unitech/pm2                                #    2s
	https://github.com/styled-components/styled-components        #    2s
	https://gitlab.freedesktop.org/wlroots/wlroots                #    2s
	https://github.com/alsa-project/alsa-lib                      #    2s
	'https://github.com/mm2/Little-CMS -name lcms2'               #    2s
	https://github.com/file/file                                  #    2s
	https://github.com/libevent/libevent                          #    2s
	https://github.com/lua/lua                                    #    2s
	https://github.com/xiph/opus                                  #    2s
	https://git.musl-libc.org/git/musl                            #    2s
	https://git.savannah.gnu.org/git/diffutils                    #    2s
	https://gitlab.freedesktop.org/libinput/libinput              #    2s
	https://github.com/libjpeg-turbo/libjpeg-turbo                #    2s
	https://github.com/libarchive/libarchive                      #    3s
	https://gitlab.freedesktop.org/xorg/lib/libx11                #    3s
	https://github.com/hakimel/reveal.js                          #    3s
	https://github.com/reduxjs/redux                              #    3s
	https://gitlab.freedesktop.org/mesa/drm                       #    3s
	https://github.com/eggert/tz                                  #    3s
	https://github.com/jquery/jquery                              #    3s
	https://github.com/scrapy/scrapy                              #    3s
	https://github.com/yarnpkg/yarn                               #    3s
	https://github.com/tiangolo/fastapi                           #    3s
	https://git.savannah.gnu.org/git/grep                         #    3s
	https://github.com/tmux/tmux                                  #    4s
	https://github.com/vuejs/vue                                  #    4s
	https://github.com/nginx/nginx                                #    4s
	https://github.com/x64dbg/x64dbg                              #    4s
	https://github.com/chartjs/Chart.js                           #    4s
	https://github.com/ReactTraining/react-router                 #    4s
	https://github.com/jekyll/jekyll                              #    4s
	https://github.com/tesseract-ocr/tesseract                    #    4s
	https://github.com/square/okhttp                              #    4s
	https://github.com/Alamofire/Alamofire                        #    4s
	https://github.com/lighttpd/lighttpd1.4                       #    4s
	https://github.com/moment/moment                              #    5s
	https://github.com/fastlane/fastlane                          #    5s
	https://gitlab.com/procps-ng/procps                           #    5s
	https://github.com/libexpat/libexpat                          #    5s
	https://github.com/d3/d3                                      #    5s
	'https://github.com/vuejs/core -name vue3'                    #    5s
	https://github.com/facebook/zstd                              #    6s
	https://gitlab.freedesktop.org/wayland/weston                 #    6s
	https://github.com/serverless/serverless                      #    6s
	https://github.com/gohugoio/hugo                              #    7s
	https://gitlab.freedesktop.org/pipewire/pipewire              #    7s
	https://github.com/nwjs/nw.js                                 #    7s
	https://github.com/apache/dubbo                               #    7s
	https://github.com/angular/angular.js                         #    8s
	https://github.com/google/guava                               #    8s
	https://gitlab.freedesktop.org/dbus/dbus                      #    8s
	https://github.com/vercel/hyper                               #    8s
	https://github.com/faker-js/faker                             #    9s
	https://git.savannah.gnu.org/git/make                         #    9s
	https://github.com/microsoft/PowerToys                        #   10s
	https://github.com/jgthms/bulma                               #   10s
	https://github.com/microsoft/terminal                         #   11s
	https://github.com/puppeteer/puppeteer                        #   11s
	https://github.com/syncthing/syncthing                        #   11s
	https://github.com/electron/electron                          #   12s
	https://github.com/ReactiveX/RxJava                           #   12s
	https://github.com/nuxt/nuxt.js                               #   12s
	https://gitlab.freedesktop.org/pulseaudio/pulseaudio          #   12s
	https://github.com/webpack/webpack                            #   13s
	https://github.com/bluez/bluez                                #   13s
	https://github.com/facebook/react                             #   14s
	https://github.com/redis/redis                                #   14s
	https://github.com/mozilla/pdf.js                             #   14s
	https://gitlab.freedesktop.org/freetype/freetype              #   14s
	https://github.com/pixijs/pixijs                              #   14s
	https://github.com/meteor/meteor                              #   16s
	https://github.com/facebook/jest                              #   16s
	https://gitlab.freedesktop.org/xorg/xserver                   #   16s
	https://github.com/prettier/prettier                          #   18s
	https://github.com/prometheus/prometheus                      #   19s
	'git://git.code.sf.net/p/libpng/code -name libpng'            #   20s
	https://github.com/rakudo/rakudo                              #   20s
	https://github.com/denoland/deno                              #   21s
	https://github.com/Semantic-Org/Semantic-UI                   #   21s
	https://github.com/gogs/gogs                                  #   21s
	https://github.com/libsdl-org/SDL                             #   22s
	https://github.com/sudo-project/sudo                          #   23s
	https://github.com/spring-projects/spring-boot                #   25s
	https://github.com/netdata/netdata                            #   25s
	https://github.com/babel/babel                                #   25s
	https://github.com/twbs/bootstrap                             #   26s
	https://github.com/django/django                              #   26s
	https://github.com/lodash/lodash                              #   26s
	https://github.com/moby/moby                                  #   27s
	https://github.com/bitcoin/bitcoin                            #   28s
	https://github.com/ansible/ansible                            #   28s
	https://github.com/scikit-learn/scikit-learn                  #   28s
	https://github.com/hashicorp/terraform                        #   30s
	https://github.com/spring-projects/spring-framework           #   31s
	https://github.com/opentofu/opentofu                          #   32s
	https://gitlab.com/gnutls/gnutls                              #   32s
	https://github.com/rails/rails                                #   33s
	https://github.com/ionic-team/ionic-framework                 #   34s
	https://github.com/apache/echarts                             #   35s
	https://gitlab.gnome.org/GNOME/gegl                           #   36s
	https://github.com/go-gitea/gitea                             #   36s
	https://github.com/zsh-users/zsh                              #   37s
	https://codeberg.org/forgejo/forgejo                          #   39s
	https://github.com/strapi/strapi                              #   40s
	https://github.com/JuliaLang/julia                            #   41s
	https://gitlab.gnome.org/GNOME/glib                           #   45s
	https://github.com/util-linux/util-linux                      #   47s
	https://github.com/angular/angular                            #   47s
	https://github.com/git/git                                    #   47s
	https://gitlab.gnome.org/GNOME/libxml2                        #   49s
	https://github.com/apache/httpd                               #   51s
	https://github.com/apache/superset                            #   54s
	https://github.com/TryGhost/Ghost                             #   55s
	https://github.com/getsentry/sentry                           #   58s
	https://github.com/sqlite/sqlite                              #   58s
	https://github.com/neovim/neovim                              #   64s
	'https://github.com/home-assistant/core -name home-assistant' #   65s
	https://git.savannah.gnu.org/git/coreutils                    #   70s
	https://github.com/ant-design/ant-design                      #   74s
	https://github.com/flutter/flutter                            #   75s
	https://github.com/opensearch-project/OpenSearch              #   75s
	https://github.com/grafana/grafana                            #   80s
	https://github.com/mpv-player/mpv                             #   84s
	https://github.com/storybookjs/storybook                      #   90s
	https://github.com/discourse/discourse                        #   94s
	https://github.com/openssl/openssl                            #   94s
	https://github.com/golang/go                                  #   95s
	https://github.com/vim/vim                                    #  101s
	https://git.ffmpeg.org/ffmpeg                                 #  102s
	https://github.com/elastic/elasticsearch                      #  103s
	https://github.com/ziglang/zig                                #  103s
	https://github.com/DragonFlyBSD/DragonFlyBSD                  #  104s
	https://github.com/microsoft/vscode                           #  110s
	https://github.com/mrdoob/three.js                            #  114s
	https://github.com/unicode-org/icu                            #  114s
	https://github.com/nestjs/nest                                #  119s
	https://github.com/nodejs/node                                #  141s
	https://git.dpkg.org/git/dpkg/dpkg                            #  153s
	https://github.com/mongodb/mongo                              #  155s
	https://github.com/apache/openoffice                          #  156s
	https://git.savannah.gnu.org/git/gawk                         #  166s
	https://git.savannah.gnu.org/git/grub                         #  174s
	https://github.com/apple/swift                                #  196s
	https://github.com/godotengine/godot                          #  196s
	https://github.com/kubernetes/kubernetes                      #  207s
	https://github.com/Perl/perl5                                 #  212s
	https://github.com/rust-lang/rust                             #  217s
	https://github.com/postgres/postgres                          #  227s
	https://gitlab.freedesktop.org/mesa/mesa                      #  239s
	https://github.com/python/cpython                             #  256s
	https://github.com/ruby/ruby                                  #  327s
	'https://github.com/openbsd/src -name OpenBSD'                #  329s
	https://github.com/microsoft/TypeScript                       #  371s
	https://github.com/JetBrains/kotlin                           #  559s
	https://gitlab.gnome.org/GNOME/gimp                           #  578s
	'https://github.com/MariaDB/server -name MariaDB'             #  584s
	'https://github.com/freebsd/freebsd-src -name FreeBSD'        #  608s
	https://gitlab.gnome.org/GNOME/gtk                            #  638s
	https://sourceware.org/git/glibc                              #  646s
	'https://github.com/NetBSD/src -name NetBSD'                  #  709s
	'https://github.com/mysql/mysql-server -name MySQL'           #  875s
	'https://git.libreoffice.org/core -name LibreOffice'          # 1022s
	https://github.com/torvalds/linux                             # 1274s
	https://sourceware.org/git/binutils-gdb                       # 1565s
	git://git.sv.gnu.org/emacs                                    # 1914s
	git://gcc.gnu.org/git/gcc                                     # 3617s
)

for r in $repos; do
	n=$r:t
	a=($=r)
	[[ $#a -gt 1 ]] && n=$a[-1]
	[[ $(psql -XtA git-stats -c "select count(*) from repos where name = '$n'") -eq 1 ]] && continue

	git-stats update -cache ./src -keep $=argv $=r
done
