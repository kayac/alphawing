$(function () {

    // msg
    var MSG = {
        PROMPT_EMAIL: '追加するメンバーのアドレスを入力してください。',
        CONFIRM_DELETE_APP: [
            'プロジェクトを削除します。',
            'このプロジェクトのすべてのファイルに、アクセスできなくなります。',
            '本当によろしいですか?'
        ].join('\n'),
        CONFIRM_DELETE_BUNDLE: 'このバージョンを削除します。よろしいですか?',
        CONFIRM_DELETE_AUTHORITY: '本当に削除しますか?',
        CONFIRM_DELETE_SELF: [
            'こちらはあなた自身のアカウントです。',
            '削除するとこのプロジェクトにアクセスできなくなります。よろしいですか?'
        ].join('\n'),
        ERROR_APP_ID: 'error:\n不正なapp idです。',
        ERROR_AUTHORITY_ID: 'error:\n不正なauthority idです。'
    };


    // touch device flags
    var ua = navigator.userAgent.toLowerCase();
    var isIOS = /iphone|ipod|ipad/.test(ua);
    var isAndroid = /android/.test(ua);
    var isTouchDevice = isIOS || isAndroid;


    // submit post
    function submitPost (href, param) {
        param = param || {};

        var $form = $('<form />');
        $form.attr({
            action: href,
            method: 'POST'
        });

        for (var key in param) {
            $form.append($('<input />').attr({
                type: 'hidden',
                name: key,
                value: param[key]
            }));
        }

        $(document.body).append($form);
        $form.submit();
    }


    // flash
    (function () {
        var $flash = $('#flash');

        $flash.hide();

        $flash.on('click', function () {
            $flash.slideUp();
        });

        setTimeout(function () {
            $flash.slideDown();
        });
    })();


    // delete app btn
    $('.btn--delete-app').on('click', function (e) {
        e.preventDefault();

        var $btn = $(this);
        var href = $btn.attr('href');

        if (!href) {
            return;
        }
        if (!window.confirm(MSG.CONFIRM_DELETE_APP)) {
            return;
        }

        submitPost(href);
    });

    // delete bundle btn
    $('.btn--delete-bundle').on('click', function (e) {
        e.preventDefault();

        var $btn = $(this);
        var href = $btn.attr('href');

        if (!href) {
            return;
        }
        if (!window.confirm(MSG.CONFIRM_DELETE_BUNDLE)) {
            return;
        }

        submitPost(href);
    });

    // authority form
    (function () {
        var $memberList = $('#member-list');
        var $memberListAdd = $('#member-list-add');

        // insert authority
        $memberListAdd.on('click', function (e) {
            e.preventDefault();

            var newEmail = window.prompt(MSG.PROMPT_EMAIL);
            if (!newEmail) {
                return;
            }

            var appId = $('#data-app-id').attr('data-app-id');
            if (!appId) {
                alert(MSG.ERROR_APP_ID);
                return;
            }

            submitPost('/app/' + appId + '/create_authority', {
                email: newEmail
            });
        });


        // delete authority
        $memberList.on('click', '.members__item__delete', function (e) {
            e.preventDefault();
            
            var $li = $(this).parent();

            var appId = $('#data-app-id').attr('data-app-id');
            var authorityId = $li.attr('data-authority-id');
            var isSelf = $li.is('.members__item--self');

            if (!appId) {
                alert(MSG.ERROR_APP_ID);
                return;
            }
            if (!authorityId) {
                alert(MSG.ERROR_AUTHORITY_ID);
                return;
            }

            if (!window.confirm(MSG.CONFIRM_DELETE_AUTHORITY)) {
                return;
            }
            if (isSelf && !window.confirm(MSG.CONFIRM_DELETE_SELF)) {
                return;
            }

            submitPost('/app/' + appId + '/delete_authority', {
                authorityId: authorityId
            });
        });
    })();


    // sp optimize
    (function () {
        if (!isTouchDevice) {
            return;
        }
        $([
            '.api-token',
            '.btn--create-app',
            '.btn--update-app',
            '.btn--delete-app',
            '.btn--create-bundle',
            '.btn--update-bundle',
            '.btn--delete-bundle',
            '.members__item--add',
            '.members__item__delete'
        ].join(',')).hide();
    })();


    // api-token form
    (function () {
        var $input = $('.api-token__token input[type="text"]');
        var value = $input.val();
        $input.on('click', function () {
            $input.select();
        });
        $input.on('change', function (e) {
            e.preventDefault();
            $input.val(value);
        });
    })();


    // bundle list tab
    (function () {
        var BUNDLE_LIST_HEIGHT = 300;
        var LABELS = ['Android', 'iOS'];
        var NAV_CLASS_NAME = 'app-detail__bundle-nav';
        var ACTIVE_CLASS_NAME = 'active';

        var $appBundle = $('#app-bundle');

        if (!$appBundle.length) {
            return;
        }

        var $nav = $('<div />');
        $nav.addClass(NAV_CLASS_NAME);
        $appBundle.after($nav);

        var selectTab = function (pos) {
            $appBundle.children().hide();
            $appBundle.children().eq(pos).show();

            $nav.children().removeClass(ACTIVE_CLASS_NAME);
            $nav.children().eq(pos).addClass(ACTIVE_CLASS_NAME);
        };

        $appBundle.css('height', BUNDLE_LIST_HEIGHT + 'px');
        $appBundle.children().css({
            position:'absolute',
            top: '0px'
        });

        $.each(LABELS, function (index, label) {
            var $btn = $('<a href="#" />');
            $btn.text(label);
            $btn.on('click', function (e) {
                e.preventDefault();
                selectTab(index);
            });
            $nav.append($btn);
        });

        // iOSの場合のみ、デフォルトでタブ1表示
        selectTab(isIOS ? 1 : 0);

        // SPの場合、タブは切り替えさせない
        if (isTouchDevice) {
            $nav.remove();
        }
    })();
});
