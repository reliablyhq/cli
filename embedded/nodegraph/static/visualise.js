export default function define(runtime, observer) {
  const main = runtime.module();
  // const fileAttachments = new Map([["suits.csv",new URL("./data.json",import.meta.url)]]);
  // main.builtin("FileAttachment", runtime.fileAttachments(name => fileAttachments.get(name)));
  
  main.variable(observer()).define(["md"], function(md){return(
    md`## Reliably Node Graph`
  )});

  main.variable(observer("chart")).define("chart", ["data","d3","width","height","types","color","location","drag","linkArc","invalidation"], 
    (data,d3,width,height,types,color,location,drag,linkArc,invalidation) => {
      const links = data.edges.map(d => Object.create(d));
      const nodes = data.nodes.map(d => Object.create(d));

      const simulation = d3.forceSimulation(nodes)
          .force("link", d3.forceLink(links).id(d => d.id))
          .force("charge", d3.forceManyBody().strength(-300))
          .force("x", d3.forceX())
          .force("y", d3.forceY())
          .force('collide', d3.forceCollide(d => 85))

      const svg = d3.create("svg")
          .attr("viewBox", [-width / 2, -height / 2, width, height])

      // Per-type markers, as they don't inherit styles.
      svg.append("defs").selectAll("marker")
          .data(types)
          .join("marker")
          .attr("id", d => `arrow-${d}`)
          .attr("viewBox", "0 -5 10 10")
          .attr("refX", 38)
          .attr("refY", 0)
          .attr("markerWidth", 6)
          .attr("markerHeight", 6)
          .attr("orient", "auto")
          .append("path")
          .attr("fill", color)
          .attr("d", 'M0,-5L10,0L0,5');

      const link = svg.append("g")
          .attr("fill", "none")
          .attr("stroke-width", 1.5)
          .selectAll("path")
          .data(links)
          .join("path")
          .attr("stroke", d => color("edge"))
          .attr("marker-end", d => `url(${new URL(`#arrow-edge`, location)})`);

      const node = svg.append("g")
          .attr("fill", "currentColor")
          .attr("stroke-linecap", "round")
          .attr("stroke-linejoin", "round")
          .selectAll("g")
          .data(nodes)
          .join("g")
          .call(drag(simulation));

      node.append("circle")
          .attr("stroke", "white")
          .attr("stroke-width", 1.5)
          .attr("r", 25)
          .attr('fill', d => '#6baed6');
    
      node.append("text")
          .attr("x", 30 + 4)
          .attr("y", "0.31em")
          .text(d=> JSON.stringify(`name: ${d.metadata.labels["name"]}`) )
          .clone(true).lower()
          .attr("fill", "none")
          .attr("stroke", "white")
          .attr("stroke-width", 3);
    
      node.on('dblclick', (e, d) => console.log(nodes[d.index]))


      simulation.on("tick", () => {
          link.attr("d", linkArc);
          node.attr("transform", d => `translate(${d.x},${d.y})`);
      });

      invalidation.then(() => simulation.stop());

      // helper function to parse type from id
      function typeFromID(id) {
        let obj = "entity_context" // default
        data.nodes.forEach((v) => { 
          if (v.id == id) { 
            obj = v.kind
          }})
        return obj
      }

      return svg.node();
  }
  );

  main.variable().define("types", ["data"], (data) => {
    // return( Array.from(new Set(data.nodes.map(d => d.kind)) )
    return (["edge"]) // hard coding the type here as it is only used to color the arrow heads
  });

  main.variable().define("data", () => {
    return ( GETSync("/data") )
    // return ( GETSync("/data.json") )
  });


  main.variable().define("height", () => { return(800) });
  main.variable().define("color", ["d3","types"], (d3,types) => {
    return(d3.scaleOrdinal(types, d3.schemeCategory10))
  });


  main.variable().define("linkArc", () => {
    return( d =>`M${d.source.x},${d.source.y}A0,0 0 0,1 ${d.target.x},${d.target.y}`) 
  });

  main.variable().define("drag", ["d3"], (d3) => {
    return(
      simulation => {
      
      function dragstarted(event, d) {
        if (!event.active) simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
      }
      
      function dragged(event, d) {
        d.fx = event.x;
        d.fy = event.y;
      }
      
      function dragended(event, d) {
        if (!event.active) simulation.alphaTarget(0);
        d.fx = null;
        d.fy = null;
      }
      
      return d3.drag()
          .on("start", dragstarted)
          .on("drag", dragged)
          .on("end", dragended);
    }
  )});

  main.variable().define("d3", ["require"], (require) => {
    return( require("d3@6") ) 
  });

  return main;
}

function GETAsync(uri, callback) {
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.onreadystatechange = function() { 
      if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
          callback(xmlHttp.responseText);
  }
  xmlHttp.open("GET", uri, true); // true for asynchronous 
  xmlHttp.send(null);
}

function GETSync(uri) {
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open( "GET", uri, false ); // false for synchronous request
  xmlHttp.send( null );
  return JSON.parse(xmlHttp.responseText);
}
