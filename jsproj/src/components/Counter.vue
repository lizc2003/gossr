<template>
  <div>
    <div>{{count}}</div>
    <div>
      <button class='btn' @click='increaseCount'> +1 </button>
      <button class='btn' @click='decreaseCount'> -1 </button>
    </div>
  </div>
</template>

<script>
  export default{
    props: {
      count: Number,
      increaseCount: Function,
      decreaseCount: Function
    }
  };
</script>

<style lang="scss">
  .btn {
    width: 100px;
    height: 60px;
    background-color: gray;
    margin: 10px;
    border-radius: 3px;
  }
</style>
