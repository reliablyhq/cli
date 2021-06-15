package report

const SLOTemplate = `# Reliably SLO Report


<style>
html {
	font-family: sans-serif;
}
table, th, td {
	border: 1px solid #ccc;
	border-collapse: collapse;
  }
td {
	padding: 5px;
}
</style>

Service Level Objectives identify what you should care about on your system. They are what good looks like for the users of your system. If an SLO is underperforming, it will be impacting your users in some way.

<details>
  <summary>Expand for further SLOs with Reliably</summary>


  The Reliably CLI allows you to define SLOs for Availability and Latency.

  An Availability SLO allows you to specify a target availability percentage for a Service.

  A Latency SLO allows you to specify a threshold latency for a service and a target percentage. The percentage gives the target percentage of responses within that threshold latency.
</details>

### Error Budget

When you define an SLO for your system, you include a target percentage for that SLO. An example target could be 95%. That leaves 5%, which is your error budget for your SLO.

<details>
  <summary>Expand for further information on Error Budgets with Reliably</summary>


  When you define an SLO with the Reliably CLI, you specify a target percentage for the Availability or Latency for that SLO.

	Over a time window, the expectation is the responses for the Service will be within that target percentage.


	An example could be 99.5% available for a period of 7 days. The target availability is less than 100%, which leaves a margin for error. The reaming 0.5% in this example can be considered the Error Budget.


</details>


For more details of an SLO report, see the Reliably documentation on [How the Reliably CLI works].


[How the Reliably CLI works]:https://reliably.com/docs/guides/how-it-works/slo-reports/

{{ $report := .Rep }}

Report time: {{ dateTime $report.Timestamp }}
{{ $reps := .Lreps }}
{{ range $index, $service := $report.Services }}
## Service #{{ serviceNo $index}}: {{$service.Name}}

| | Name      | Current | Objective| Time Window | Type  | Trend |
|-|------------| ---:| ---:|---:|----|:--:|
{{ range $ind, $sl := $service.ServiceLevels -}}
	|{{- svcLevelGetStatusIcon $sl -}}
	|{{- svcLevelGetName $sl}}|
	{{- svcLevelGetActualResult $sl}}|
	{{- svcLevelGetObjective $sl }}|
	{{- svcLevelGetTimeWindow $sl }}|
	{{- svcLevelGetType $sl }}|
	{{- svcLevelGetTrends $service.Name $sl $reps }}
{{ end }}



The Error Budget metrics are:

|  Type    | Name          |ErrorBudget(%)|Time Window|Downtime|Consumed|Remain
|---|-----------|---:|---:|---:|---:|---:|
{{ range $ind, $sl := $service.ServiceLevels -}}
|{{- svcLevelGetType $sl -}}|
{{- svcLevelGetName $sl}}|
{{- svcLevelGetTimeWindow $sl }}|
{{- errBudgetPercentage $sl }}|
{{- errBudgetAllowedDownTime $sl }}|
{{- errBudgetConsumed $sl }}|
{{- errBudgetRemain $sl }}|
{{ end }}


{{ end }}

<small>Generated with: The Reliably CLI Version {{ reliablyVersion }}</small>

`
