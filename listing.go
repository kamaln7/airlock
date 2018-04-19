package airlock

import (
	"bytes"
	"fmt"
	"html/template"
)

const HtmlFileListingTemplate = `
<!DOCTYPE html>
<html lang="en">
  <title>{{if .IsNotRoot}}/{{.Name}}{{else}}{{.Name}}{{end}}</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="robots" content="noindex">
  <style>
  /* tachyons */html{line-height:1.15;-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}body{margin:0}header,nav{display:block}h1{font-size:2em;margin:.67em 0}main{display:block}a{background-color:transparent;-webkit-text-decoration-skip:objects}/* 1 */::-webkit-file-upload-button{-webkit-appearance:button;font:inherit}/* 1 */a,body,h1,header,html,li,main,nav,ul{box-sizing:border-box}.bb{border-bottom-style:solid;border-bottom-width:1px}.b--white-50{border-color:hsla(0,0%,100%,.5)}.bw1{border-width:.125rem}.sans-serif{font-family:-apple-system,BlinkMacSystemFont,avenir next,avenir,helvetica neue,helvetica,ubuntu,roboto,noto,segoe ui,arial,sans-serif}.fw4{font-weight:400}.link{text-decoration:none}.link,.link:active,.link:focus,.link:hover,.link:link,.link:visited{transition:color .15s ease-in}.link:focus{outline:1px dotted currentColor}.list{list-style-type:none}.white-90{color:hsla(0,0%,100%,.9)}.white{color:#fff}.bg-white{background-color:#fff}.pb3{padding-bottom:1rem}.pv3{padding-top:1rem;padding-bottom:1rem}.ph4{padding-left:2rem;padding-right:2rem}.f3{font-size:1.5rem}.f4{font-size:1.25rem}.dim{opacity:1}.dim,.dim:focus,.dim:hover{transition:opacity .15s ease-in}.dim:focus,.dim:hover{opacity:.5}.dim:active{opacity:.8;transition:opacity .15s ease-out}
  .break-wrap { overflow-wrap: break-word; }
  .lh-tall { line-height: 2; }
  .spaces { color: #11126b; }
  .bg-spaces { background-color: #11126b; }
  </style>
  <body class="sans-serif bg-spaces white">
    <header class="bg-white spaces ph4 pv3">
        <h1 class="f3 break-wrap"><span class="fw4">Contents of</span> {{if .IsNotRoot}}/{{.Name}}{{else}}/{{end}}</h1>
    </header>
    <main class="ph4 pv3">
        <nav>
            <ul class="lh-tall list f4">
                {{if .IsNotRoot}}
					<li class="pb3">
						<a href="../index.html" class="link white-90 dim f4">←back</a>
					</li>
				{{end}}
				{{range .Children}}
                	{{if .IsNotRoot}}
						<li>{{if .IsDir}}📂{{else}}📄{{end}} <a href="./{{.Name}}{{if .IsDir}}/index.html{{end}}" class="link white-90 bw1 bb b--white-50">{{.Name}}</a></li>
					{{end}}
				{{end}}
            </ul>
        </nav>
    </main>
  </body>
</html>
`

func (a *Airlock) AddFileListings() error {
	tmpl, err := template.New("index").Parse(HtmlFileListingTemplate)
	if err != nil {
		return err
	}
	a.listingTmpl = tmpl

	a.addFileListing(a.tree["."], true)
	return nil
}

func (a *Airlock) addFileListing(tree *File, root bool) {
	if !tree.IsDir {
		return
	}

	// don't want to recursively go into the root dir
	if tree.RelPath == "." && !root {
		return
	}

	if len(tree.Children) == 0 {
		return
	}

	// check if an index file already exists
	hasIndex := false
	for _, child := range tree.Children {
		if child.Name == "index.html" {
			hasIndex = true
			break
		}
	}

	if !hasIndex {
		var relPath string
		if tree.RelPath == "." {
			relPath = "index.html"
		} else {
			relPath = fmt.Sprintf("%s/index.html", tree.RelPath)
		}

		index := &File{
			RelPath: relPath,
			Name:    "index.html",
			IsDir:   false,
			Read:    a.makeListingReader(tree),
		}

		a.tree[index.RelPath] = index
		tree.Children = append(tree.Children, index)
		a.files = append(a.files, index)
	}

	for _, child := range tree.Children {
		a.addFileListing(child, false)
	}
}

func (a *Airlock) makeListingReader(tree *File) FileReader {
	return func() ([]byte, error) {
		var out bytes.Buffer
		err := a.listingTmpl.Execute(&out, tree)
		if err != nil {
			return nil, err
		}

		return out.Bytes(), nil
	}
}
