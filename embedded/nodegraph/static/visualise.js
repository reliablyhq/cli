export default function define(runtime, observer) {
  const main = runtime.module();
  // const fileAttachments = new Map([["suits.csv",new URL("./data.json",import.meta.url)]]);
  // main.builtin("FileAttachment", runtime.fileAttachments(name => fileAttachments.get(name)));

  main.variable(observer("chart")).define("chart", ["data","d3","width","height","types","color","location","drag","linkArc","invalidation"], 
    (data,d3,width,height,types,color,location,drag,linkArc,invalidation) => {
      const links = data.edges.map(d => Object.create(d));
      const nodes = data.nodes.map(d => Object.create(d));

      const simulation = d3.forceSimulation(nodes)
          .force("link", d3.forceLink(links).id(d => d.id))
          .force("charge", d3.forceManyBody().strength(-300))
          .force("x", d3.forceX())
          .force("y", d3.forceY())
          .force('collide', d3.forceCollide(d => 85)) // collide, the forced distance between nodes

      const svg = d3.create("svg")
          .attr("viewBox", [-width / 2, -height / 2, width, height]);

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

      let labels = [];
      let labelsObjects = [];
      // [
      //  { name: property1, values: [value1, value2] },
      //  { name: property2, values: [value3, value4] },
      // ]

      const circle = node.append("circle")
          .attr("class", (d) => {
            let objectiveLabels = d.metadata.labels;
            let classString = "node-circle";
            for (const property in objectiveLabels) {
              if (property !== "name") {
                if (labels.includes(property)) {
                  const propertyValues = labelsObjects[labels.indexOf(property)];
                  if (propertyValues.values.includes(objectiveLabels[property])) {
                    classString = classString.concat(` group-${property}-${propertyValues.values.indexOf(objectiveLabels[property])}`);
                  } else {
                    propertyValues.values.push(objectiveLabels[property]);
                    classString = classString.concat(` group-${property}-${propertyValues.values.indexOf(objectiveLabels[property])}`);
                  }
                } else {
                  labels.push(property);
                  labelsObjects.push({ name: property, values: [objectiveLabels[property]] });
                  classString = classString.concat(` group-${property}-0`);
                }
              }
            }
            return classString;
            // let p = JSON.stringify(`${d.metadata.labels["purpose"]}`).substring(1).slice(0, -1);
            // if (purposes.includes(p)) {
            //   return `node-circle group-${purposes.indexOf(p)}`;
            // } else {
            //   purposes.push(p)
            //   return `node-circle group-${purposes.length - 1}`;
            // }
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
            let str = d.metadata.labels["name"];
            return str;
          });

      label.append("text")
          .attr("class", "node-metadata")

          .attr("x", 34)
          .attr("y", 13)
          // .data((d) => { d.metadata.labels })
          // .join(tspan)
          //     .attr("x", 34)
          //     .attr("dy", "1.2em")
          //     .text((d) => { d });

          .text((d) => {
            let str = "";
            for (const property in d.metadata.labels) {
              str = str.concat(`${property}: ${d.metadata.labels[property]} - `);
            }
            str = str.slice(0, -3);
            return str;
            // if (d.metadata.labels["purpose"] !== undefined) {
            //   let str = JSON.stringify(`${d.metadata.labels["purpose"]}`).substring(1).slice(0, -1);
            //   return str;
            // } else {
            //   return "-";
            // }
          });

      // node.on('dblclick', (e, d) => console.log(nodes[d.index]))

      circle.on('mouseover',(e, d) => {
        d3.select(e.target.parentNode).raise().classed("display-tooltip", true);
        let label = e.target.parentNode.childNodes.item(1);
        let width = label.getBBox().width + 4;
        let rect = label.childNodes.item(0);
        rect.setAttribute("width", width);
      });

      circle.on('mouseout', (e) => {
        d3.select(e.target.parentNode).classed("display-tooltip", false);
      });

      simulation.on("tick", () => {
          link.attr("d", linkArc);
          node.attr("transform", d => `translate(${d.x},${d.y})`);
      });

      let form = d3.select("#js-group-select-form");

      form.append("label")
              .attr("class", "group-select__label")
              .attr("for", "group-select__input")
              .text("Highlight nodes by label");

      let selectInput = form.append("select")
              .attr("class", "group-select__input")
              .attr("id", "group-select")
              .attr("name", "group-select")

      selectInput.append("option")
          .text("Select a label")
          .attr("value", "");

      selectInput.selectAll("options")
          .data(labels)
          .enter()
              .append("option")
                  .text((d) => { return d; })
                  .attr("value", (d) => { return d; });

      let defaultColor = {
        fill: "hsl(170, 25%, 60%)",
        stroke: "hsl(170, 50%, 35%)"
      };

      let colors = [
        {
          fill: "hsl(260, 25%, 60%)",
          stroke: "hsl(260, 50%, 35%)"
        },
        {
          fill: "hsl(350, 25%, 60%)",
          stroke: "hsl(350, 50%, 35%)"
        },
        {
          fill: "hsl(80, 25%, 60%)",
          stroke: "hsl(80, 50%, 35%)"
        },
        {
          fill: "hsl(200, 25%, 60%)",
          stroke: "hsl(200, 50%, 35%)"
        },
        {
          fill: "hsl(290, 25%, 60%)",
          stroke: "hsl(290, 50%, 35%)"
        },
        {
          fill: "hsl(20, 25%, 60%)",
          stroke: "hsl(20, 50%, 35%)"
        },
        {
          fill: "hsl(110, 25%, 60%)",
          stroke: "hsl(110, 50%, 35%)"
        },
        {
          fill: "hsl(230, 25%, 60%)",
          stroke: "hsl(230, 50%, 35%)"
        },
        {
          fill: "hsl(320, 25%, 60%)",
          stroke: "hsl(320, 50%, 35%)"
        },
        {
          fill: "hsl(50, 25%, 60%)",
          stroke: "hsl(50, 50%, 35%)"
        },
      ];

      selectInput.on("change", (e) => {
        d3.selectAll(".node-circle")
            .attr("fill", defaultColor.fill)
            .attr("stroke", defaultColor.stroke);
        let selectedLabel = e.target.value;
        let length = labelsObjects[labels.indexOf(selectedLabel)].values.length;
        for (let i = 0; i < length; i++) {
          d3.selectAll(`.group-${selectedLabel}-${i}`)
              .attr("fill", colors[i % length].fill)
              .attr("stroke", colors[i % length].stroke);
        }
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

function removeQuotes(string) {
  return string.substring(1).slice(0, -1);
}
