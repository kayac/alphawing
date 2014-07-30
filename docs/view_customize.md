# View Customize

If you want to modify the appearance of alphawing, you need to build JavaScript and CSS.

| Path                        | Roll                    |
|:----------------------------|:------------------------|
| `/static/_src`              | front-end source root   |
| `/static/_src/sass`         | Sass(.scss) files       |
| `/static/_src/js`           | JavaScript source files |
| `/static/_src/Gruntfile.js` | Grunt config file       |

## Get Dependencies

You should get [Compass](http://compass-style.org/) & [Grunt](http://gruntjs.com/) for front-end development.

```sh
# Grunt
$ node -v
$ npm install -g grunt-cli

# Compass (Sass)
$ ruby -v
$ gem install compass

# npm modules
$ cd static/_src
$ npm install
```

## Build

Run `grunt` in order to build SCSS and optimize JavaScript.

```sh
$ cd staitc/_src
$ grunt
```

You can use the `watch` task that automatically runs build on the file changes under `/static/_src`.

```sh
$ cd staitc/_src
$ grunt watch
```
