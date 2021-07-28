import React, { Component } from "react";

import { color } from "chart.js/helpers";
import ChartComponent from "react-chartjs-2";


class Heatmap extends Component {
  render() {
    if (this.props.data === undefined) {
      return null;
    }

    let data = {
      labels: [],
      datasets: [
        {
          label: "Heatmap",
          data: [],
          minValue: -1,
          maxValue: -1,
          xSet: {},
          ySet: {},
          fromColor: this.props.fromColor.match(/\d+/g).map(Number),
          toColor: this.props.toColor.match(/\d+/g).map(Number),
          backgroundColor: ctx => {
            if (ctx.dataset === undefined || ctx.dataset.data === undefined || ctx.dataset.data[ctx.dataIndex] === undefined) {
              return color('white').rgbString();
            }

            const value = ctx.dataset.data[ctx.dataIndex].v;
            const steps = ctx.dataset.maxValue - ctx.dataset.minValue + 1;
            const step = value - ctx.dataset.minValue;
            const factor = 1 / steps * step;

            let result = ctx.dataset.fromColor.slice();
            for (var i = 0; i < 3; i++) {
                result[i] = Math.round(result[i] + factor * (ctx.dataset.toColor[i] - ctx.dataset.fromColor[i]));
            }

            return color(result).rgbString();
          },
          borderWidth: 0,
          width: ctx => {
            return (ctx.chart.chartArea || {}).width / Object.keys(ctx.dataset.xSet).length - 1;
          },
          height: ctx => {
            return (ctx.chart.chartArea || {}).height / Object.keys(ctx.dataset.ySet).length - 1;
          },
        },
      ],
    };

    let options = {
      animation: false,
      maintainAspectRatio: false,
      scales: {
        y: {
          type: "category",
          offset: true,
          grid: {
            display: false,
          },
        },
        x: {
          type: "time",
          offset: true,
          labels: [],
          grid: {
            display: false,
          },

        },
      },
      plugins: {
        legend: false,
        tooltip: {
          callbacks: {
            title: () => {
              return '';
            },
            label: ctx => {
              const v = ctx.dataset.data[ctx.dataIndex].v;
              return 'Count: ' + v;
            },
          },
        },
      },
    };

    for (const row of this.props.data) {
      options.scales.x.labels.push(row.x);
      data.datasets[0].xSet[row.x] = {};

      Object.entries(row.y).forEach(([k, v]) => {
        data.datasets[0].ySet[k] = {};

        data.datasets[0].data.push({
          x: row.x,
          y: k,
          v: v,
        });

        if (data.datasets[0].minValue === -1 || data.datasets[0].minValue > v) {
          data.datasets[0].minValue = v;
        }

        if (data.datasets[0].maxValue < v) {
          data.datasets[0].maxValue = v;
        }
      });
    }

    return(
      <ChartComponent type="matrix" data={data} options={options} />
    );
  }
}

export default Heatmap;
