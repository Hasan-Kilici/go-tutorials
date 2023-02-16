function openBox(id){
    document.getElementById(id).style.display = "block";
}

function closeBox(id){
    document.getElementById(id).style.display = "none";
}
window.onload = ()=>{
    let repos;
        fetch("/github/hasan-kilici/repositories").then(async(data)=>{
            repos = await data.json();
        })
        setTimeout(()=>{
            for(let i = 0;i < repos.length;i++){
                    document.getElementById("repos").innerHTML += `
                    <div class="col-md-4 mt-2 mb-2">
                      <div class="card bg-dark text-light">
                        <div class="card-body">
                            ${repos[i].name}
                        </div>
                        <div class="card-footer d-flex gap-2 align-items-center">
                          <div><i class="fa-solid fa-circle text-primary"></i> ${repos[i].language}</div>
                          <div><i class="fa-solid fa-star"></i> ${repos[i].stargazers_count}</div>
                        </div>
                      </div>
                    </div>
                    `;
            }
        },1000)
}