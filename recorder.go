/*
===========================================================================
AXIS GRABBER GPL Source Code
Copyright (C) 2011 Vasileios Anagnostopoulos.
This file is part of the AXIS GRABBER GPL Source Code (?AXIS GRABBER GPL Source Code?).  
AXIS GRABBER GPL Source Code is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
AXIS GRABBER GPL Source Code is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with AXIS GRABBER GPL Source Code.  If not, see <http://www.gnu.org/licenses/>.
In addition, the AXIS GRABBER GPL Source Code is also subject to certain additional terms. You should have received a copy of these additional terms immediately following the terms and conditions of the GNU General Public License which accompanied the Doom 3 Source Code.  If not, please request a copy in writing from id Software at the address below.
If you have questions concerning this license or the applicable additional terms, you may contact in writing Vasileios Anagnostopoulos, Campani 3 Street, Athens Greece, POBOX 11252.
===========================================================================
*/

package main
import (
    "fmt"
    "http"
    "os"
    "mime"
    "mime/multipart"
    "strconv"
    "syscall"
    "flag"
    "time"
)
var camera_name *string = flag.String("camname", "null", "camera name")
var camera_ip *string = flag.String("camaddress", "null", "camera ip address")
var camera_port *int = flag.Int("camport", -1 , "camera port")
var save_folder *string = flag.String("savefolder","null", "folder to save frames")


// MultipartReader returns a MIME multipart reader if this is a
// multipart/form-data POST request, else returns nil and an error.
func MultipartReader(rr * http.Response) (multipart.Reader, os.Error) {
    v, ok := rr.Header["Content-Type"]
    if !ok {
        return nil, http.ErrNotMultipart
    }
    d, params := mime.ParseMediaType(v)
    if d != "multipart/x-mixed-replace" {
        return nil, http.ErrNotMultipart
    }
    boundary, ok := params["boundary"]
    if !ok {
        return nil, http.ErrMissingBoundary
    }
    boundary="myboundary"
    fmt.Println(boundary)
    return multipart.NewReader(rr.Body, boundary), nil
}

func helpuser() {
    fmt.Println("you should use the program as")
    fmt.Println("program --camname=<camera name> --camaddress=<ip address> --camport=<camera port> --savefolder=<folder to save frames>")

}

func parseME() bool {
    flag.Parse()
    if (*camera_name == "null") {
        fmt.Println("No camera name in arguments")
        helpuser()
        return false
    }
    if (*camera_ip == "null") {
        fmt.Println("No camera ip address in arguments")
        helpuser()
        return false
    }
    if (*camera_port == -1 ) {
        fmt.Println("No camera port in arguments")
        helpuser()
        return false
    }
    if ( *save_folder == "null") {
        fmt.Println("No folder to save in arguments")
        helpuser()
        return false
    }
    return true
}

func main(){
    var rr * http.Response;
    var myreader multipart.Reader

    syscall.Umask(0000)

    if( !parseME()){
        os.Exit(1);
    }

    var requesturl string;

    requesturl= ("http://"+ (*camera_ip) + ":" + strconv.Itoa(*camera_port)+ "/mjpg/1/video.mjpg")
    fmt.Println("request sent to "+requesturl)
    rr,_,_=http.Get(requesturl)
    myreader,_=MultipartReader(rr)

    var p * multipart.Part

    var curr_length int=0;

    var templen int
    var buff []byte
    var s string
    var m int;

    var info *os.FileInfo
    var err os.Error

    info,err=os.Lstat(*save_folder)
    if(err!=nil){
       fmt.Println("Folder "+ (*save_folder) + " Is problematic")
       fmt.Println(err.String())
       os.Exit(1)
    }

    if(!info.IsDirectory()){
       fmt.Println("Folder "+ (*save_folder) + " Is not a directory")
       os.Exit(1)
    }

    var foldertime *time.Time=nil;
    var foldersecs int64
    var folderstamp string

    var tstamp_secs int64;
    var tstamp_nsecs int64;

    var msecs int64
    var update bool
    var foldername string
    var imagename string
    var mywriter *os.File


    for i:=0;i<1; {

        p,_=myreader.NextPart()

        update=false

        tstamp_secs,tstamp_nsecs,_=os.Time()
        if(foldertime==nil) {
            foldertime=time.SecondsToLocalTime(tstamp_secs)
            foldersecs=tstamp_secs
            update=true
        }  else {
           if(tstamp_secs > foldersecs) {
               foldertime=time.SecondsToLocalTime(tstamp_secs)
               foldersecs=tstamp_secs
               update=true
           }

        }

        if ( update) {
            folderstamp=strconv.Itoa64(foldertime.Year)+"_"+
                strconv.Itoa(foldertime.Month)+"_"+
                strconv.Itoa(foldertime.Day)+"_"+
                strconv.Itoa(foldertime.Hour)+"_"+
                strconv.Itoa(foldertime.Minute)+"_"+
                strconv.Itoa(foldertime.Second)
            foldername=(*save_folder)+"/"+ (*camera_name)+"_"+folderstamp
            err=os.Mkdir(foldername, 0700)
            if(err != nil) {
            fmt.Fprintf(os.Stderr, "error creating %s because : %s\n", foldername , err.String())
            os.Exit(1)
            }
        }

        templen,_=strconv.Atoi(p.Header["Content-Length"])

        if(templen > curr_length) {
            curr_length=templen
            buff = make([]byte,curr_length)
        }

        for counter:=0 ;  counter < templen ; {
           m,_=p.Read(buff[counter:templen])
           counter+=m;
        }

        p.Close()
        msecs= tstamp_nsecs / 1e6
        imagename= "image_"+ folderstamp + "_" + strconv.Itoa64(msecs)+".jpg"
        s=foldername + "/" + imagename

        mywriter,err=os.Open(s,os.O_CREAT | os.O_WRONLY,0600)

        if(err != nil) {
            fmt.Fprintf(os.Stderr, "error writing %d bytes because : %s\n", templen , err.String())
            os.Exit(1)
        }

        for counter:=0 ;  counter < templen ; {
           m,_=mywriter.Write(buff[counter : templen])
           counter+=m;
        }
   }
}
