package chromedpundetected

const (
	scriptWebDriverAttr = `
    Object.defineProperty(window, 'navigator', {
        value: new Proxy(navigator, {
                has: (target, key) => (key === 'webdriver' ? false : key in target),
                get: (target, key) =>
                        key === 'webdriver' ?
                        false :
                        typeof target[key] === 'function' ?
                        target[key].bind(target) :
                        target[key]
                })
    });
			`

	// Scripts from : https://github.com/microlinkhq/browserless/tree/master/packages/goto/src/evasions
	scriptHeadless = `
    Object.defineProperty(navigator, 'maxTouchPoints', {get: () => 1});
    Object.defineProperty(navigator.connection, 'rtt', {get: () => 100});

    // https://github.com/microlinkhq/browserless/blob/master/packages/goto/src/evasions/chrome-runtime.js
    window.chrome = {
        app: {
            isInstalled: false,
            InstallState: {
                DISABLED: 'disabled',
                INSTALLED: 'installed',
                NOT_INSTALLED: 'not_installed'
            },
            RunningState: {
                CANNOT_RUN: 'cannot_run',
                READY_TO_RUN: 'ready_to_run',
                RUNNING: 'running'
            }
        },
        runtime: {
            OnInstalledReason: {
                CHROME_UPDATE: 'chrome_update',
                INSTALL: 'install',
                SHARED_MODULE_UPDATE: 'shared_module_update',
                UPDATE: 'update'
            },
            OnRestartRequiredReason: {
                APP_UPDATE: 'app_update',
                OS_UPDATE: 'os_update',
                PERIODIC: 'periodic'
            },
            PlatformArch: {
                ARM: 'arm',
                ARM64: 'arm64',
                MIPS: 'mips',
                MIPS64: 'mips64',
                X86_32: 'x86-32',
                X86_64: 'x86-64'
            },
            PlatformNaclArch: {
                ARM: 'arm',
                MIPS: 'mips',
                MIPS64: 'mips64',
                X86_32: 'x86-32',
                X86_64: 'x86-64'
            },
            PlatformOs: {
                ANDROID: 'android',
                CROS: 'cros',
                LINUX: 'linux',
                MAC: 'mac',
                OPENBSD: 'openbsd',
                WIN: 'win'
            },
            RequestUpdateCheckStatus: {
                NO_UPDATE: 'no_update',
                THROTTLED: 'throttled',
                UPDATE_AVAILABLE: 'update_available'
            }
        }
    }

    // https://github.com/microlinkhq/browserless/blob/master/packages/goto/src/evasions/navigator-permissions.js
    if (!window.Notification) {
        window.Notification = {
            permission: 'denied'
        }
    }

    const originalQuery = window.navigator.permissions.query
    window.navigator.permissions.__proto__.query = parameters =>
        parameters.name === 'notifications'
            ? Promise.resolve({ state: window.Notification.permission })
            : originalQuery(parameters)

    const oldCall = Function.prototype.call
    function call() {
        return oldCall.apply(this, arguments)
    }
    Function.prototype.call = call

    const nativeToStringFunctionString = Error.toString().replace(/Error/g, 'toString')
    const oldToString = Function.prototype.toString

    function functionToString() {
        if (this === window.navigator.permissions.query) {
            return 'function query() { [native code] }'
        }
        if (this === functionToString) {
            return nativeToStringFunctionString
        }
        return oldCall.call(oldToString, this)
    }
    // eslint-disable-next-line
    Function.prototype.toString = functionToString
	`
)
