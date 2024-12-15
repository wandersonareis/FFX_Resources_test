import { Injectable } from '@angular/core';
import {Compress} from "../../wailsjs/go/services/CompressService";

@Injectable({
  providedIn: 'root'
})
export class CompressService {
  async compress(file: string) {
    Compress(file);
  }
}
