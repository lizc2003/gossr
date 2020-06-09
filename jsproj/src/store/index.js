import Vue from 'vue';
import Vuex from 'vuex';

Vue.use(Vuex);

export function createStore (context) {
    return new Vuex.Store({
        state: {
            meta: {
                Title: "SSR demo",
                Keywords: "ssr vue",
                Description: "This is a ssr demo",
                OgImage: "https://github.githubassets.com/images/modules/site/logos/google-logo.png",
            },
            count: 0,
            pageData: {},
        },
        mutations: {
            increaseCount (state) {
                state.count++
            },
            decreaseCount (state) {
                state.count--
            },
        }
    })
}
