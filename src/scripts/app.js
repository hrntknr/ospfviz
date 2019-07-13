import axios from 'axios';
import * as d3 from 'd3';
import {Netmask} from 'netmask';

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

const tooltip = d3.select('body').append('div').attr('class', 'tooltip');

axios.get('/api/ospf').then(({data: routers})=>{
  const links = [
  ];
  const nodes = [
  ];
  const drs = {};
  const p2p = {};

  routers.forEach((router)=>{
    nodes.push({
      isRouter: true,
      isInterface: false,
      routerID: router.advRouter,
      links: router.links,
      // hostname: router.hostname,
    });
  });

  routers.forEach((router, routerIndex)=>{
    router.links.forEach((link)=>{
      switch (link.type) {
      case 1: {
        if (p2p[`${link.link.neighbor}-${nodes[routerIndex].routerID}`] == null) {
          p2p[`${nodes[routerIndex].routerID}-${link.link.neighbor}`] = routerIndex;
        } else {
          const source1 = nodes.length;
          nodes.push({
            isRouter: false,
            isInterface: true,
          });
          links.push({source: source1, target: routerIndex, isInterfaceLink: true});

          const source2 = nodes.length;
          nodes.push({
            isRouter: false,
            isInterface: true,
          });
          links.push({source: source1, target: source2});

          const target = p2p[`${link.link.neighbor}-${nodes[routerIndex].routerID}`];
          links.push({source: source2, target, isInterfaceLink: true});
        }
        break;
      }
      case 2: {
        if (drs[link.link.dr] == null) {
          drs[link.link.dr] = nodes.length;
          nodes.push({
            isRouter: false,
            isInterface: false,
          });
        }
        const source = nodes.length;
        nodes.push({
          isRouter: false,
          isInterface: true,
        });
        links.push({source, target: drs[link.link.dr]});
        links.push({source, target: routerIndex, isInterfaceLink: true});
        drs[link.link.dr];
        break;
      }
      case 3: {
        const source = nodes.length;
        nodes.push({
          isRouter: false,
          isInterface: false,
        });
        links.push({source, target: routerIndex});
        break;
      }
      }
    });
  });

  const simulation = d3.forceSimulation(nodes)
    .force('link', d3.forceLink(links))
    .force('link', d3.forceLink(links).id((d)=>d.id).distance(0).strength((link)=>link.isInterfaceLink ? 3 : 1))
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
    .call(drag(simulation))
    .on('mouseover', (d)=>{
      if (d.isRouter) {
        let html = `<p>${d.routerID}</p>`;
        // if (d.hostname.length != 0) html+=d.hostname.map((hostname)=>`<p>${hostname}</p>`).join();
        if (d.links.length != 0) {
          html+=`Links<br>${d.links
            .filter((link)=>link.type==3)
            .map((link)=>`${link.link.network}/${link.link.mask}`)
            .join('<br>')}`;
        }
        tooltip
          .style('visibility', 'visible')
          .html(html);
      }
    })
    .on('mousemove', (d)=>{
      if (d.isRouter) {
        tooltip
          .style('top', (d3.event.pageY - 20) + 'px')
          .style('left', (d3.event.pageX + 10) + 'px');
      }
    })
    .on('mouseout', (d)=>{
      if (d.isRouter) {
        tooltip.style('visibility', 'hidden');
      }
    });

  simulation.on('tick', ()=>{
    link
      .attr('x1', (d)=>d.source.x)
      .attr('y1', (d)=>d.source.y)
      .attr('x2', (d)=>d.target.x)
      .attr('y2', (d)=>d.target.y);
    node
      .attr('cx', (d)=>d.x)
      .attr('cy', (d)=>d.y);
  });

  document.getElementById('app').appendChild(svg.node());
});
