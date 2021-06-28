export default function define(runtime, observer) {
  const main = runtime.module();
  // const fileAttachments = new Map([["suits.csv",new URL("./data.json",import.meta.url)]]);
  // main.builtin("FileAttachment", runtime.fileAttachments(name => fileAttachments.get(name)));
  
  main.variable(observer()).define(["md"], function(md){return(
    md`# Reliably Node Graph`
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
          .attr("viewBox", [-width / 2, -height / 2, width, height - 64]);

      // Per-type markers, as they don't inherit styles.
      svg.append("defs").selectAll("marker")
          .data(types)
          .join("marker")
          .attr("id", d => `arrow-${d}`)
          .attr("viewBox", "0 -5 10 10")
          .attr("refX", 38)
          .attr("refY", 0)
          .attr("markerWidth", 9)
          .attr("markerHeight", 9)
          .attr("orient", "auto")
          .append("path")
          .attr("fill", "hsl(170, 5%, 48%)")
          .attr("d", 'M0,-5L10,0L0,5');

      const link = svg.append("g")
          .attr("fill", "none")
          .attr("stroke-width", 1)
          .selectAll("path")
          .data(links)
          .join("path")
          .attr("stroke", "hsl(170, 5%, 48%)")
          .attr("marker-end", d => `url(${new URL(`#arrow-edge`, location)})`);

      const node = svg.append("g")
          .attr("class", "nodes")
          .attr("fill", "currentColor")
          .attr("stroke-linecap", "round")
          .attr("stroke-linejoin", "round")
          .selectAll("g")
          .data(nodes)
          .join("g")
          .attr("class", "node")
          .call(drag(simulation));

      let purposes = [];

      const circle = node.append("circle")
          .attr("class", (d) => {
            let p = JSON.stringify(`${d.metadata.labels["purpose"]}`).substring(1).slice(0, -1);
            if (purposes.includes(p)) {
              return `node-circle group-${purposes.indexOf(p)}`;
            } else {
              purposes.push(p)
              return `node-circle group-${purposes.length - 1}`;
            }
          })
          .attr("stroke", "hsl(170, 40%, 35%)")
          .attr("stroke-width", 1.5)
          .attr("r", 25)
          .attr("fill", "hsl(170, 25%, 60%)");

      var label = node.append("g")
          .attr("class", "node-label");
      
      label.append("rect")
          .attr("class", "node-label-bg")
          .attr("width", 1)
          .attr("height", 32)
          .attr("fill", "none")
          .attr("x", 30)
          .attr("y", -14);

      label.append("text")
          .attr("class", "node-name")
          .attr("x", 34)
          .attr("y", 2)
          .text((d) => {
            let str = JSON.stringify(`${d.metadata.labels["name"]}`).substring(1).slice(0, -1);
            return str;
          });
      
      label.append("text")
          .attr("class", "node-purpose")
          .attr("x", 34)
          .attr("y", 13)
          .text((d) => {
            if (d.metadata.labels["purpose"] !== undefined) {
              let str = JSON.stringify(`${d.metadata.labels["purpose"]}`).substring(1).slice(0, -1);
              return str;
            } else {
              return "-";
            }
          });

      // node.on('dblclick', (e, d) => console.log(nodes[d.index]))

      circle.on('mouseover',(e, d) => {
        let label = e.target.parentNode.childNodes.item(1);
        let width = label.getBBox().width + 4;
        let rect = label.childNodes.item(0);
        rect.setAttribute("width", width);
      });

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
