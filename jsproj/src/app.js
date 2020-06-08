import Vue from 'vue';
import App from './App.vue';
import { createRouter } from './router'
import { createStore } from './store'
import axios from 'axios';

export function createApp (context) {
  const store = createStore(context);
  const router = createRouter();

  if (typeof window == 'undefined') {
    if(context && context.ssrCtx) {
      axios.defaults.headers.common['SSR-Ctx'] = context.ssrCtx;
    }
  }

  const app = new Vue({
      router,
      store,
      render: h => h(App)
    });

  return { app, router, store }
}

export function callComponentsHookWith(compoList, hookName, context) {
  return compoList.map((component) => {
      const hook = component[hookName];
      if (hook) {
          return hook(context);
      }
  }).filter(_ => _);
}
