<template lang="pug">
  .app
    .number-info
      .machine.card
        .card-title 在线采集器数量:
        counting(:num="8", background="darkkhaki")
      .collection-quantity.card
        .card-title 采集数据总数:
        counting(:num="total", background="lightskyblue")
    .total-change.card
      .card-title 采集总量变化趋势:
      Chart(v-model="totalChartData")
    .efficiency-change.card
      .card-title 采集效率变化趋势:
      Chart(v-model="efficiencyChartData")
</template>

<script>
import 'echarts'
import Chart from 'echarts-middleware'
import counting from '@puge/scoreboard'
const axios = require('axios')
export default {
  name: 'app',
  components: {
    Chart,
    counting
  },
  data () {
    return {
      total: 0,
      totalChartData: {
        xAxis: {
          type: 'category',
          data: []
        },
        yAxis: {
          type: 'value',
          scale: true,
          axisLabel: {
            formatter: '{value} 条'
          }
        },
        series: [{
          data: [],
          areaStyle: {},
          type: 'line'
        }]
      },
      efficiencyChartData: {
        xAxis: {
          type: 'category',
          data: []
        },
        yAxis: {
          type: 'value',
          axisLabel: {
            formatter: '{value} 条/秒'
          }
        },
        series: [{
          data: [],
          type: 'line',
          lineStyle: {
            color: '#009fe9'
          },
          markLine: {
            data: [
              {type: 'average', name: '平均值'}
            ]
          }
        }]
      }
    }
  },
  mounted () {
    Array.prototype.queue = function(val, num) {
      if (this.length < num) {
        this.push(val)
      } else {
        this.splice(0, 1)
        this.push(val)
      }
    }
    this.updata()
    // 每5秒刷新一次数据
    setInterval(this.updata, 5000)
  },
  methods: {
    updata: function () {
      axios.get('http://127.0.0.1:8200/').then((res) => {
        const myDate = new Date()
        const value = res.data
        // 第一次统计效率恒为0
        if (this.total === 0) this.total = value.total
        this.totalChartData.xAxis.data.queue(myDate.getHours() + ":" + myDate.getMinutes() + ':' + myDate.getSeconds(), 720)
        this.totalChartData.series[0].data.queue(value.total, 720)
        this.efficiencyChartData.xAxis.data.queue(myDate.getHours() + ":" + myDate.getMinutes() + ':' + myDate.getSeconds(), 720)
        this.efficiencyChartData.series[0].data.queue((value.total - this.total) / 5, 720)
        this.total = value.total
      })
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
}
html, body, .app {
  height: 100%;
  width: 100%;
  background-color: #f3f3f3;
}
.app {
  max-width: 1200px;
  margin: 0 auto;
}
.card {
  margin: 5px;
  border-radius: 3px;
  overflow: hidden;
  background-color: white;
  box-shadow: 1px 1px 11px #bfbcbc;
}
.board {
  margin: 30px 0;
}
.card-title {
  background-color: teal;
  color: white;
  line-height: 30px;
  padding: 0 5px;
}
.number-info {
  display: flex;
}
.collection-quantity {
  height: 200px;
  width: calc(100% - 320px);
}
.total-change.card, .efficiency-change {
  height: 400px;
}
.machine {
  width: 300px;
}
</style>
