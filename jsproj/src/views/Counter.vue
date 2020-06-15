<template lang="html">
  <div>
    <Counter :count='count' :increase-count='increaseCount' :decrease-count='decreaseCount'></Counter>
    <br /><br />
    <p>XMLHttpRequest result:</p>
    <div>{{xhrResult}}</div>
  </div>
</template>

<script>
  import Counter from '../components/Counter.vue';
  import { mapState, mapMutations } from 'vuex';

  export default {
    asyncData({store, route, context}) {
      return store.dispatch('xhrTest')
    },
    components: {
      Counter
    },
    computed: {
      ...mapState([
        'count',
        'xhrResult',
      ])
    },
    methods: {
      ...mapMutations([
        'increaseCount',
        'decreaseCount',
      ])
    }
  };
</script>
