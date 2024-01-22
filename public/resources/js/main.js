const log = console.log,
    ref = Vue.ref,
    toRef = Vue.toRef,
    toRefs = Vue.toRefs,
    isProxy = Vue.isProxy,
    createApp = Vue.createApp,
    reactive = Vue.reactive,
    isArray = Array.isArray,
    mergeRefs = (...refs) => {
        let _ref = {};
        for (const ref of refs) {
            for (const key in ref) {
                _ref[key] = toRef(ref, key);
            }
        }
        return _ref;
    },
    apiURL = 'http://localhost:81',
    request = (route, setup) => {
        if (!setup) setup = {}
        if (!setup.headers) setup.headers = {};
        setup.headers["Content-Type"] = "application/json";
        return fetch(apiURL + route, setup).then(r => r.json())
    },
    get = route => request(route),
    post = (route, data) => request(route, {
        method: 'POST',
        body: JSON.stringify(data)
    });

const header = reactive({
    title: '代理更新',
    subTitle: 'Hello',
});

const copyright = reactive({
    org: '随心记事',
    year: new Date().getFullYear(),
    autoUpdate() {
        setInterval(_ => {
            this.year = new Date().format('yy:MM:dd HH:mm:ss')
        }, 1000);
        return this;
    },
    newYear(y) {
        this.year = y;
    }
}).autoUpdate();


const footer = reactive({
    copyright,
})

const pager = mergeRefs(header, footer, copyright)
let app = createApp({
    setup() {
        return pager
    },
    components: {
        ...ElementPlusIconsVue,
    }
}).use(ElementPlus, {
    locale: ElementPlusLocaleZhCn,
}).mount('#app');
