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

function int2ip(ipInt) {
  return ( (ipInt>>>24) +'.' + (ipInt>>16 & 255) +'.' + (ipInt>>8 & 255) +'.' + (ipInt & 255) );
}

const tooltip = d3.select('body').append('div').attr('class', 'tooltip');

axios.get('/api/ospf/v2').then(({data: routers})=>{
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
      router: router,
      advRouter: router.advRouter,
      links: router.contents,
      // hostname: router.hostname,
    });
  });

  routers.forEach((router, routerIndex)=>{
    router.contents.forEach((link)=>{
      switch (link.type) {
      case 1: {
        if (p2p[`${link.link.linkID}-${nodes[routerIndex].advRouter}`] == null) {
          p2p[`${nodes[routerIndex].advRouter}-${link.link.linkID}`] = routerIndex;
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

          const target = p2p[`${link.link.linkID}-${nodes[routerIndex].advRouter}`];
          links.push({source: source2, target, isInterfaceLink: true});
        }
        break;
      }
      case 2: {
        if (drs[link.link.linkID] == null) {
          drs[link.link.linkID] = nodes.length;
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
        links.push({source, target: drs[link.link.linkID]});
        links.push({source, target: routerIndex, isInterfaceLink: true});
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
      default: {
        console.log(`Unknown NetworkType ${link.type}`);
        break;
      }
      }
    });
  });

  const simulation = d3.forceSimulation(nodes)
    .force('link', d3.forceLink(links).distance(0).strength((link)=>link.isInterfaceLink ? 3 : 1))
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
        const html = `<p>${int2ip(d.advRouter)}</p>`;
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

  document.getElementById('ospfv2').appendChild(svg.node());
});


axios.get('/api/ospf/v3').then(({data})=>{
  const links = [
  ];
  const nodes = [
  ];
  const drs = {};
  const p2p = {};
  const advRouters = {};
  const routers = data.filter((router)=>router.lsType == 0x2001);
  const stubs = data.filter((router)=>router.lsType == 0x2009);

  routers.forEach((router)=>{
    if (router.lsType == 0x2001) {
      const routerIndex = nodes.length;
      nodes.push({
        isRouter: true,
        isInterface: false,
        advRouter: router.advRouter,
        links: router.contents,
      // hostname: router.hostname,
      });
      advRouters[router.advRouter] = routerIndex;
    }
  });

  routers.forEach((router, routerIndex)=>{
    router.contents.forEach((link)=>{
      switch (link.type) {
      case 1: {
        if (p2p[`${link.link.neighborADVRouter}-${nodes[routerIndex].advRouter}`] == null) {
          p2p[`${nodes[routerIndex].advRouter}-${link.link.neighborADVRouter}`] = routerIndex;
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

          const target = p2p[`${link.link.neighborADVRouter}-${nodes[routerIndex].advRouter}`];
          links.push({source: source2, target, isInterfaceLink: true});
        }
        break;
      }
      case 2: {
        if (drs[`${link.link.neighborADVRouter}-${link.link.neighborInterfaceID}`] == null) {
          drs[`${link.link.neighborADVRouter}-${link.link.neighborInterfaceID}`] = nodes.length;
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
        links.push({source, target: drs[`${link.link.neighborADVRouter}-${link.link.neighborInterfaceID}`]});
        links.push({source, target: routerIndex, isInterfaceLink: true});
        drs[link.link.dr];
        break;
      }
      default: {
        console.log(`Unknown NetworkType ${link.type}`);
        break;
      }
      }
    });
  });

  stubs.forEach((stub)=>{
    const source = nodes.length;
    nodes.push({
      isRouter: false,
      isInterface: false,
    });
    links.push({source, target: advRouters[stub.advRouter]});
  });

  const simulation = d3.forceSimulation(nodes)
    .force('link', d3.forceLink(links).distance(0).strength((link)=>link.isInterfaceLink ? 3 : 1))
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
        const html = `<p>${int2ip(d.advRouter)}</p>`;
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

  document.getElementById('ospfv3').appendChild(svg.node());
});
