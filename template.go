package main

const bmTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>bm</title>
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap-theme.min.css">
<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/js/bootstrap.min.js"></script>
<style type="text/css">
    .bm{
    	margin: 20px;
    }
</style>
</head>
<body>
<div class="bm">
    <table class="table table-hover">
        <thead>
            <tr>
                <th>URL</th>
                <th>Last modified</th>
            </tr>
        </thead>
        <tbody>
			<tr>
					{{ $map := .Bmap }}
					{{ with .Sorted }}
						{{range $key := .}}
							{{ with $map }}
								{{ $value := index $map $key }}
								<tr>
									<td> <a href={{$value.URL}}>{{ $value.Title }}</a></td>
									<td> {{ humanize $value.Modified }}</td>
								</tr>
								{{ end }}
							{{end}}
					{{end}}
			</tr>
        </tbody>
    </table>
</div>
</body>
</html>
`
