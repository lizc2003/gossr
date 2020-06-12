<template lang="html">
  <div>
    <Counter :count='count' :increase-count='increaseCount' :decrease-count='decreaseCount'></Counter>
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
        'count'
      ])
    },
    methods: {
      ...mapMutations([
        'increaseCount',
        'decreaseCount'
      ])
    }
  };
</script>
