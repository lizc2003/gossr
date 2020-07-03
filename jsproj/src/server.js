import { createApp, callComponentsHookWith } from './app.js'

export default (context) => {
  return new Promise((resolve, reject) => {
    const { app, router,store } = createApp(context)

    router.push(context.url)
    router.onReady(() => {
      try {
        const matchedComponents = router.getMatchedComponents()
        if (!matchedComponents.length) {
          return reject({code: 404})
        }

        callComponentsHookWith(matchedComponents, 'prepareVuex', {store})
        Promise.all(callComponentsHookWith(matchedComponents, 'asyncData',
            {store, route: router.currentRoute, context})
        ).then(() => {
          context.state = store.state
          context.meta = context.state.meta
          context.state.meta = {}

          resolve(app)
        }).catch((err) => {
          reject(err)
        });
      } catch(err) {
        reject(err)
      }
    }, (err) => {
      reject(err)
    })
  })
}
