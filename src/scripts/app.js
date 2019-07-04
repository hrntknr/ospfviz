import axios from 'axios';
import * as d3 from 'd3';

const width = 1920;
const height = 1080;

function drag(simulation) {
  function dragstarted(d) {
    if (!d3.event.active) simulation.alphaTarget(0.3).restart();
    d.fx = d.x;
    d.fy = d.y;
  }
  function dragged(d) {
    d.fx = d3.event.x;
    d.fy = d3.event.y;
  }
  function dragended(d) {
    if (!d3.event.active) simulation.alphaTarget(0);
    d.fx = null;
    d.fy = null;
  }
  return d3.drag()
    .on('start', dragstarted)
    .on('drag', dragged)
    .on('end', dragended);
}

axios.get('/api/ospf').then(({data: routers})=>{
  console.log(routers);
  const links = [
  ];
  const nodes = [
  ];
  const drs = {};
  const p2p = {};

  routers.forEach((router)=>{
    nodes.push({
      data: router,
      isRouter: true,
      label: router.routerID,
    });
  });

  routers.forEach((router, routerIndex)=>{
    router.links.forEach((link)=>{
      switch (link.type) {
      case 0: {
        const source = nodes.length;
        nodes.push({
          data: link.stub,
        });
        links.push({source, target: routerIndex});
        break;
      }
      case 1: {
        if (drs[link.transit.dr] == null) {
          drs[link.transit.dr] = nodes.length;
          nodes.push({
            data: link.transit.dr,
          });
        }
        const source = nodes.length;
        nodes.push({
          data: link,
          isInterface: true,
        });
        links.push({source, target: drs[link.transit.dr]});
        links.push({source, target: routerIndex});
        drs[link.transit.dr];
        break;
      }
      case 2: {
        if (p2p[`${link.p2p.neighbor}-${nodes[routerIndex].data.routerID}`] == null) {
          p2p[`${nodes[routerIndex].data.routerID}-${link.p2p.neighbor}`] = routerIndex;
        } else {
          const source = p2p[`${link.p2p.neighbor}-${nodes[routerIndex].data.routerID}`];
          links.push({source, target: routerIndex});
        }
        break;
      }
      default: {
        break;
      }
      }
    });
  });

  const simulation = d3.forceSimulation(nodes)
    .force('link', d3.forceLink(links))
    .force('link', d3.forceLink(links).id((d) => d.id).distance(0).strength(1.5))
    .force('charge', d3.forceManyBody().strength(-50))
    .force('x', d3.forceX())
    .force('y', d3.forceY());

  const svg = d3.create('svg')
    .attr('viewBox', [-width / 2, -height / 2, width, height]);

  const link = svg.append('g')
    .attr('stroke', '#999')
    .attr('stroke-opacity', 0.6)
    .selectAll('line')
    .data(links)
    .join('line');

  const node = svg.append('g')
    .attr('stroke-width', 1.5)
    .selectAll('circle')
    .data(nodes)
    .join('circle')
    .attr('r', (d)=>d.isRouter ? 10 : 3.5)
    .attr('fill', (d)=>d.isInterface || d.isRouter ? '#fff' : '#000')
    .attr('stroke', (d)=>d.isInterface || d.isRouter ? '#000' : '#fff')
    .call(drag(simulation));

  const text = node.append('text')
    .attr('dx', 20)
    .attr('dy', 0)
    .text((d)=>d.label ? d.label : '');

  simulation.on('tick', () => {
    link
      .attr('x1', (d)=>d.source.x)
      .attr('y1', (d)=>d.source.y)
      .attr('x2', (d)=>d.target.x)
      .attr('y2', (d)=>d.target.y);
    node
      .attr('cx', (d)=>d.x)
      .attr('cy', (d)=>d.y);
    text
      .attr('x', (d)=>d.x)
      .attr('y', (d)=>d.y);
  });

  document.getElementById('app').appendChild(svg.node());
});
