const width = 800;
const height = 800;
// nodeの定義。ここを増やすと楽しい。
const nodes = [
  {id: 0, label: 'nodeA'},
  {id: 1, label: 'nodeB'},
  {id: 2, label: 'nodeC'},
  {id: 3, label: 'nodeD'},
  {id: 4, label: 'nodeE'},
  {id: 5, label: 'nodeF'},
];

// node同士の紐付け設定。実用の際は、ここをどう作るかが難しいのかも。
const links = [
  {source: 0, target: 1},
  {source: 0, target: 2},
  {source: 1, target: 3},
  {source: 1, target: 3},
  {source: 2, target: 1},
  {source: 2, target: 3},
  {source: 3, target: 4},
  {source: 4, target: 5},
  {source: 5, target: 3},
];
// forceLayout自体の設定はここ。ここをいじると楽しい。
const force = d3.layout.force()
  .nodes(nodes)
  .links(links)
  .size([width, height])
  .distance(140) // node同士の距離
  .friction(0.9) // 摩擦力(加速度)的なものらしい。
  .charge(-100) // 寄っていこうとする力。推進力(反発力)というらしい。
  .gravity(0.1) // 画面の中央に引っ張る力。引力。
  .start();

// svg領域の作成
const svg = d3.select('body')
  .append('svg')
  .attr({width: width, height: height});

// link線の描画(svgのline描画機能を利用)
const link = svg.selectAll('line')
  .data(links)
  .enter()
  .append('line')
  .style({'stroke': '#ccc',
    'stroke-width': 1,
  });

// nodesの描画(今回はsvgの円描画機能を利用)
const node = svg.selectAll('circle')
  .data(nodes)
  .enter()
  .append('circle')
  .attr({
    // せっかくなので半径をランダムに
    r: function() {
      return Math.random() * (40 - 10) + 10;
    },
  })
  .style({
    fill: 'orange',
  })
  .call(force.drag);

// nodeのラベル周りの設定
const label = svg.selectAll('text')
  .data(nodes)
  .enter()
  .append('text')
  .attr({
    'text-anchor': 'middle',
    'fill': 'white',
    'font-size': '9px',
  })
  .text(function(data) {
    return data.label;
  });

// tickイベント(力学計算が起こるたびに呼ばれるらしいので、座標追従などはここで)
force.on('tick', function() {
  link.attr({
    x1: function(data) {
      return data.source.x;
    },
    y1: function(data) {
      return data.source.y;
    },
    x2: function(data) {
      return data.target.x;
    },
    y2: function(data) {
      return data.target.y;
    },
  });
  node.attr({
    cx: function(data) {
      return data.x;
    },
    cy: function(data) {
      return data.y;
    },
  });
  // labelも追随するように
  label.attr({
    x: function(data) {
      return data.x;
    },
    y: function(data) {
      return data.y;
    },
  });
});
