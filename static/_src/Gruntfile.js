module.exports = function (grunt) {
    var path = require('path');
    var config = {};

    var alias = {
        $: 'jquery',
        _: 'underscore'
    };

    // basic
    {
        config.pkg =  grunt.file.readJSON('package.json');

        grunt.loadNpmTasks('grunt-contrib-watch');
        grunt.loadNpmTasks('grunt-contrib-copy');
        config.watch = {};
        config.copy = {};
    }

    // release
    {
        grunt.loadNpmTasks('grunt-release');

        config.release = {
            options: {
                file: 'bower.json',
                npm: false
            }
        };
    }

    var configureEnv = function (name, env) {
        // js
        {
            grunt.loadNpmTasks('grunt-auto-deps');
            config.auto_deps = config.auto_deps || {};

            config.auto_deps[name] = {
                scripts: ['alphawing'],
                dest: path.resolve(env.sitePath, 'js'),
                loadPath: ['js/*.js', 'js/lib/*.js'],
                ignore: [],
                forced: [],
                wrap: true,
                alias: alias
            };

            if (env.watch) {
                config.watch.js = {
                    files: ['js/*.js'],
                    tasks: ['auto_deps:' + name]
                };
            }    
            env.tasks.push('auto_deps:' + name);
        }


        // js lib copy
        (function () {
            var libs = [
                'bower_components/html5shiv/src/html5shiv.js'
            ];

            var libDir = path.resolve(env.sitePath, 'js') + '/lib/';
            var files = [];
            libs.forEach(function (lib) {
                files.push({
                    expand: true,
                    flatten: true,
                    src: lib,
                    dest: libDir
                    });
            });
            config.copy[name] = { files: files };
            env.tasks.push('copy:' + name);
        })();
    
    
        // css
        {
            grunt.loadNpmTasks('grunt-contrib-compass');
    
            config.compass = config.compass || {};
            config.compass[name] = {
                options: {
                    sassDir                 : 'sass',
                    cssDir                  : path.resolve(env.sitePath, 'css'),
                    javascriptsDir          : path.resolve(env.sitePath, 'js'),
                    imagesDir               : path.resolve(env.sitePath, 'img'),
                    fontsDir                : path.resolve(env.sitePath, 'fonts'),
                    generatedImagesPath     : path.resolve(env.sitePath, 'img'),
                    httpImagesPath          : path.resolve(env.httpPath, 'img'),
                    httpGeneratedImagesPath : path.resolve(env.httpPath, 'img'),
                    httpFontsPath           : path.resolve(env.httpPath, 'fonts'),
                    environment             : 'production',
                    outputStyle             : 'compressed'
                }
            };
    
            if (env.watch) {
                config.watch.css = {
                    files: ['sass/*.scss', 'sass/**/*.scss'],
                    tasks: ['compass:' + name]
                };
            }
    
            env.tasks.push('compass:' + name);
        }
    
    
        // server
        {
            grunt.loadNpmTasks('grunt-koko');
    
            config.koko = config.koko || {};
            config.koko[name] = {
                root: path.resolve(env.sitePath, path.relative(env.httpPath, '/')),
                openPath: env.httpPath
            };
    
            grunt.registerTask('server', ['koko:' + name]);
        }

        // set as task
        grunt.registerTask(name, env.tasks);
    };

    configureEnv('dev', {
        tasks: [],
        sitePath: '../',
        httpPath: '/static/',
        watch: true,
        ejs: true,
        test: true
    });

    // init
    grunt.initConfig(config);
    grunt.registerTask('default', ['dev']);
};
