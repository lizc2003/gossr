import Vue from 'vue';
import Vuex from 'vuex';
import axios from 'axios';

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
            xhrResult: "",
        },
        mutations: {
            increaseCount (state) {
                state.count++
            },
            decreaseCount (state) {
                state.count--
            },
            setXhrResult(state, result) {
                state.xhrResult = result
            }
        },
        actions: {
            xhrTest ({commit}) {
                return axios.get('/api/check').then(res => {
                    console.log("XMLHttpRequest ok.")
                    commit('setXhrResult', res.data)
                })
            }
        }
    })
}
